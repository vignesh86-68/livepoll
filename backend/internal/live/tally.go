package live

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vignesh/livepoll/internal/models"
)

// incrementScript safely increments multiple vote counts and publishes the
// updated tally in a single atomic operation.
const incrementScript = `
local pollCode = ARGV[1]
local totalIncr = tonumber(ARGV[2])

for i = 3, #ARGV, 2 do
    redis.call("HINCRBY", KEYS[1], ARGV[i], tonumber(ARGV[i+1]))
end
redis.call("INCRBY", KEYS[2], totalIncr)
local seq = redis.call("INCR", KEYS[3])

local countsFlat = redis.call("HGETALL", KEYS[1])
local total = redis.call("GET", KEYS[2])

local counts = {}
for i = 1, #countsFlat, 2 do
    counts[countsFlat[i]] = tonumber(countsFlat[i+1])
end

local tally = {
    counts = counts,
    total = tonumber(total),
    seq = tonumber(seq)
}

local event = {
    type = "tally",
    code = pollCode,
    seq = tonumber(seq),
    tally = tally,
    at = tonumber(redis.call("TIME")[1]) * 1000
}

local eventJSON = cjson.encode(event)
redis.call("PUBLISH", KEYS[4], eventJSON)

return eventJSON
`

// IncrementAndPublish runs the Lua script to update the live tally and notify viewers.
func (l *Live) IncrementAndPublish(ctx context.Context, pollID, code string, optionIDs []string) (*models.Tally, error) {
	keys := []string{
		"poll:" + pollID + ":counts",
		"poll:" + pollID + ":total",
		"poll:" + pollID + ":seq",
		"poll:" + pollID + ":events",
	}

	args := []interface{}{code, 1}
	for _, opt := range optionIDs {
		args = append(args, opt, 1)
	}

	res, err := l.client.Eval(ctx, incrementScript, keys, args...).Result()
	if err != nil {
		return nil, err
	}

	var event models.Event
	if err := json.Unmarshal([]byte(res.(string)), &event); err != nil {
		return nil, err
	}
	return event.Tally, nil
}

// GetTally retrieves the current tally from Redis.
func (l *Live) GetTally(ctx context.Context, pollID string) (*models.Tally, error) {
	countsFlat, err := l.client.HGetAll(ctx, "poll:"+pollID+":counts").Result()
	if err != nil {
		return nil, err
	}

	if len(countsFlat) == 0 {
		return nil, redis.Nil // Indicate a miss, so caller can SeedTally
	}

	counts := make(map[string]int64)
	for k, v := range countsFlat {
		counts[k], _ = strconv.ParseInt(v, 10, 64)
	}

	totalStr, _ := l.client.Get(ctx, "poll:"+pollID+":total").Result()
	total, _ := strconv.ParseInt(totalStr, 10, 64)

	seqStr, _ := l.client.Get(ctx, "poll:"+pollID+":seq").Result()
	seq, _ := strconv.ParseInt(seqStr, 10, 64)

	return &models.Tally{
		Counts: counts,
		Total:  total,
		Seq:    seq,
	}, nil
}

// SeedTally populates Redis with durable numbers from MongoDB if they were lost.
func (l *Live) SeedTally(ctx context.Context, pollID string, counts map[string]int64, total int64) error {
	pipe := l.client.Pipeline()
	if len(counts) > 0 {
		args := []interface{}{}
		for k, v := range counts {
			args = append(args, k, v)
		}
		pipe.HSet(ctx, "poll:"+pollID+":counts", args...)
	} else {
		// Just to ensure the key exists for HGETALL later
		pipe.HSet(ctx, "poll:"+pollID+":counts", "_seed", 0)
		pipe.HDel(ctx, "poll:"+pollID+":counts", "_seed")
	}
	pipe.Set(ctx, "poll:"+pollID+":total", total, 0)
	_, err := pipe.Exec(ctx)
	return err
}

// PublishClose broadcasts that a poll has closed.
func (l *Live) PublishClose(ctx context.Context, pollID, code string) error {
	event := models.Event{
		Type:   models.EventClosed,
		Code:   code,
		Status: models.StatusClosed,
		At:     time.Now().UnixMilli(),
	}
	data, _ := json.Marshal(event)
	return l.client.Publish(ctx, "poll:"+pollID+":events", data).Err()
}
