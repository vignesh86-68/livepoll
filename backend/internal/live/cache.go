package live

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CachePoll stores serialized poll data in Redis.
func (l *Live) CachePoll(ctx context.Context, code string, data []byte, ttl time.Duration) error {
	return l.client.Set(ctx, "poll:"+code+":meta", data, ttl).Err()
}

// GetCachedPoll retrieves serialized poll data from Redis.
func (l *Live) GetCachedPoll(ctx context.Context, code string) ([]byte, error) {
	data, err := l.client.Get(ctx, "poll:"+code+":meta").Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// InvalidatePoll removes a poll from the cache.
func (l *Live) InvalidatePoll(ctx context.Context, code string) error {
	return l.client.Del(ctx, "poll:"+code+":meta").Err()
}
