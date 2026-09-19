package httpapi

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/vignesh/livepoll/internal/auth"
	"github.com/vignesh/livepoll/internal/config"
	"github.com/vignesh/livepoll/internal/live"
	"github.com/vignesh/livepoll/internal/store"
	"github.com/vignesh/livepoll/internal/ws"
)

func NewRouter(cfg *config.Config, st *store.Store, l *live.Live, h *ws.Hub) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(RealIP())

	tm := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL)
	isProd := cfg.IsProduction()

	authHandler := NewAuthHandler(st, tm, l, isProd)
	pollHandler := NewPollHandler(st, l, isProd, cfg.VoterSalt)
	wsHandler := NewWSHandler(h, st, l)
	healthHandler := NewHealthHandler(st, l)

	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/ws/polls/:code", wsHandler.HandleWS)

	api := r.Group("/api")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/signup", RateLimit(l, "signup", 10), authHandler.Signup)
			authGroup.POST("/login", RateLimit(l, "login", 20), authHandler.Login)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.GET("/me", AuthRequired(tm), authHandler.Me)
		}

		pollsGroup := api.Group("/polls")
		{
			pollsGroup.POST("", AuthRequired(tm), RateLimit(l, "poll_create", 10), pollHandler.Create)
			pollsGroup.GET("/mine", AuthRequired(tm), pollHandler.ListMine)
			
			pollsGroup.GET("/:code", AuthOptional(tm), pollHandler.GetPoll)
			pollsGroup.GET("/:code/results", pollHandler.GetResults)
			pollsGroup.POST("/:code/vote", AuthOptional(tm), RateLimit(l, "vote", 60), pollHandler.Vote)
			
			pollsGroup.POST("/:code/close", AuthRequired(tm), pollHandler.Close)
			pollsGroup.DELETE("/:code", AuthRequired(tm), pollHandler.Delete)
		}
	}

	// Serve Static Files
	r.NoRoute(func(c *gin.Context) {
		path := filepath.Join(cfg.StaticDir, c.Request.URL.Path)
		if _, err := os.Stat(path); err == nil {
			c.File(path)
			return
		}
		
		// Fallback for SPA routing
		indexPath := filepath.Join(cfg.StaticDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	return r
}
