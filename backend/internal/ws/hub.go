package ws

import (
	"github.com/vignesh/livepoll/internal/live"
)

// Hub maintains the set of active rooms and broadcasts messages to the rooms.
type Hub struct {
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
	live       *live.Live
}

// NewHub creates a new Hub.
func NewHub(live *live.Live) *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		live:       live,
	}
}

// Run starts the main loop of the Hub.
//
// It is driven purely by channels, meaning no mutexes are required to safely
// mutate the rooms map.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			room, ok := h.rooms[client.pollID]
			if !ok {
				room = newRoom(client.pollID, client.code, h.live)
				h.rooms[client.pollID] = room
				go room.runSubscription()
			}
			room.clients[client] = true

		case client := <-h.unregister:
			room, ok := h.rooms[client.pollID]
			if ok {
				if _, ok := room.clients[client]; ok {
					delete(room.clients, client)
					close(client.send)
					if len(room.clients) == 0 {
						room.cancel()
						delete(h.rooms, client.pollID)
					}
				}
			}
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
