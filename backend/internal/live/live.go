package live

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Live wraps the Redis client to provide realtime features.
type Live struct {
	client *redis.Client
}

// Connect dials Redis and verifies the connection before returning.
func Connect(ctx context.Context, url string) (*Live, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("live: parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("live: ping: %w", err)
	}

	return &Live{client: client}, nil
}

// Ping reports whether Redis is reachable.
func (l *Live) Ping(ctx context.Context) error {
	return l.client.Ping(ctx).Err()
}

// Client returns the underlying Redis client, used by the WebSocket layer
// to create pub/sub subscriptions for live event fan-out.
func (l *Live) Client() *redis.Client {
	return l.client
}

// Close disconnects the client.
func (l *Live) Close() error {
	return l.client.Close()
}
