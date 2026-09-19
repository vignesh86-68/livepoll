package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vignesh/livepoll/internal/live"
	"github.com/vignesh/livepoll/internal/models"
	"github.com/vignesh/livepoll/internal/store"
	"github.com/vignesh/livepoll/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub   *ws.Hub
	store *store.Store
	live  *live.Live
}

func NewWSHandler(h *ws.Hub, s *store.Store, l *live.Live) *WSHandler {
	return &WSHandler{hub: h, store: s, live: l}
}

func (h *WSHandler) HandleWS(c *gin.Context) {
	code := c.Param("code")

	p, err := h.store.Polls.FindByCode(c.Request.Context(), code)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(h.hub, conn, p.ID.Hex(), p.Code)
	h.hub.Register(client)

	// Send initial snapshot immediately
	viewers, _ := h.live.IncrViewers(context.Background(), p.ID.Hex(), p.Code, 1)
	
	tally, _ := h.live.GetTally(context.Background(), p.ID.Hex())
	if tally == nil {
		tally = &models.Tally{Counts: p.Totals, Total: p.TotalVotes, Seq: 0}
	}

	snap := models.Event{
		Type:    models.EventSnapshot,
		Code:    p.Code,
		Tally:   tally,
		Viewers: viewers,
		Status:  p.EffectiveStatus(time.Now()),
		At:      time.Now().UnixMilli(),
	}
	
	data, _ := json.Marshal(snap)
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_ = conn.WriteMessage(websocket.TextMessage, data)

	// Decrement viewers on disconnect
	go func() {
		// Wait for client loops to finish naturally before decrementing viewers
		defer h.live.IncrViewers(context.Background(), p.ID.Hex(), p.Code, -1)
		client.ReadPump()
	}()

	go client.WritePump()
}
