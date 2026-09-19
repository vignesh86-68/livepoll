package live

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vignesh/livepoll/internal/models"
)

// IncrViewers adjusts the viewer count for a poll and publishes the new value.
func (l *Live) IncrViewers(ctx context.Context, pollID, code string, delta int64) (int64, error) {
	key := "poll:" + pollID + ":viewers"
	viewers, err := l.client.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, err
	}

	// Avoid negative counts if logic gets out of sync
	if viewers < 0 {
		viewers = 0
		l.client.Set(ctx, key, 0, 0)
	}

	event := models.Event{
		Type:    models.EventViewers,
		Code:    code,
		Viewers: viewers,
		At:      time.Now().UnixMilli(),
	}
	data, _ := json.Marshal(event)
	_ = l.client.Publish(ctx, "poll:"+pollID+":events", data).Err()

	return viewers, nil
}

// GetViewers retrieves the current viewer count for a poll.
func (l *Live) GetViewers(ctx context.Context, pollID string) (int64, error) {
	key := "poll:" + pollID + ":viewers"
	val, err := l.client.Get(ctx, key).Int64()
	if err != nil {
		return 0, nil // treat not found or err as 0 viewers
	}
	return val, nil
}
