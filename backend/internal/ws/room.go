package ws

import (
	"context"

	"github.com/vignesh/livepoll/internal/live"
)

// Room represents a poll's live connection state.
//
// When the first client joins, a Redis subscription is started.
// When the last client leaves, it is cancelled.
type Room struct {
	pollID  string
	code    string
	clients map[*Client]bool
	cancel  context.CancelFunc
	live    *live.Live
}

func newRoom(pollID, code string, l *live.Live) *Room {
	return &Room{
		pollID:  pollID,
		code:    code,
		clients: make(map[*Client]bool),
		live:    l,
	}
}

// runSubscription connects to Redis pub/sub and broadcasts messages to the room.
func (r *Room) runSubscription() {
	// A dedicated context allows cancellation when the room is empty.
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	pubsub := r.live.Client().Subscribe(ctx, "poll:"+r.pollID+":events")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			data := []byte(msg.Payload)
			// Send to all clients in the room
			for client := range r.clients {
				select {
				case client.send <- data:
				default:
					// Buffer full: slow client is lagging, drop it to avoid stalling others.
					// We do not modify r.clients directly here because runSubscription
					// runs concurrently with Hub's main loop; we let the unregister flow handle it.
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
