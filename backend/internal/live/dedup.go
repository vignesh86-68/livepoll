package live

import (
	"context"
	"time"
)

// CheckVote uses Redis SET NX as a fast path filter against double voting.
// This is not the source of truth (MongoDB is) but it drops the common case
// of duplicate clicks instantly.
func (l *Live) CheckVote(ctx context.Context, pollID, voterKey string) (bool, error) {
	key := "vote:" + pollID + ":" + voterKey
	ok, err := l.client.SetNX(ctx, key, 1, 24*time.Hour).Result()
	return ok, err
}

// ClearVote deletes the deduplication marker.
// This is called as a compensating action if the MongoDB insert fails,
// ensuring a transient error doesn't lock the voter out for 24 hours.
func (l *Live) ClearVote(ctx context.Context, pollID, voterKey string) error {
	key := "vote:" + pollID + ":" + voterKey
	return l.client.Del(ctx, key).Err()
}
