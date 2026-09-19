package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vignesh/livepoll/internal/config"
	"github.com/vignesh/livepoll/internal/httpapi"
	"github.com/vignesh/livepoll/internal/live"
	"github.com/vignesh/livepoll/internal/store"
	"github.com/vignesh/livepoll/internal/ws"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("boot: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("connecting to MongoDB...")
	st, err := store.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer st.Close(context.Background())

	if err := st.EnsureIndexes(ctx); err != nil {
		log.Fatalf("mongo index: %v", err)
	}

	log.Println("connecting to Redis...")
	lv, err := live.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis connect: %v", err)
	}
	defer lv.Close()

	log.Println("starting WS hub...")
	hub := ws.NewHub(lv)
	go hub.Run()

	r := httpapi.NewRouter(cfg, st, lv, hub)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("exiting")
}
