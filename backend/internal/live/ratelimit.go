package live

import (
	"context"
	"fmt"
	"time"
)

// Allow checks if a request is within the rate limit.
func (l *Live) Allow(ctx context.Context, scope, identifier string, limit int) (bool, error) {
	// Simple fixed window rate limit using minute resolution
	minute := time.Now().Truncate(time.Minute).Unix()
	
	key := fmt.Sprintf("rl:%s:%s:%d", scope, identifier, minute)

	pipe := l.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 60*time.Second)
	
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	
	return incr.Val() <= int64(limit), nil
}
