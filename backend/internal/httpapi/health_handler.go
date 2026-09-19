package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vignesh/livepoll/internal/live"
	"github.com/vignesh/livepoll/internal/store"
)

type HealthHandler struct {
	store *store.Store
	live  *live.Live
}

func NewHealthHandler(s *store.Store, l *live.Live) *HealthHandler {
	return &HealthHandler{store: s, live: l}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.store.Ping(ctx); err != nil {
		respondError(c, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	if err := h.live.Ping(ctx); err != nil {
		respondError(c, http.StatusServiceUnavailable, "redis unavailable")
		return
	}

	c.String(http.StatusOK, "OK")
}
