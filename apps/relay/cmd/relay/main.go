package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
	"github.com/francogalfre/coop/apps/relay/internal/transport/httpapi"
)

const (
	sweepInterval = 5 * time.Minute
	sweepMaxAge   = 30 * time.Minute
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("relay: %v", err)
	}

	ctx := context.Background()

	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("relay: %v", err)
	}
	defer pool.Close()

	if err := pool.Migrate(ctx); err != nil {
		log.Fatalf("relay: %v", err)
	}

	registry := presence.New()
	store := stream.New()
	mailbox := stream.NewMailbox()
	hub := stream.NewPresenceHub()
	takeover := stream.NewTakeoverRegistry()
	ptyHub := stream.NewPtyHub()
	steerRequests := stream.NewSteerRequestRegistry()

	go func() {
		ticker := time.NewTicker(sweepInterval)
		defer ticker.Stop()

		for now := range ticker.C {
			registry.Sweep(now, sweepMaxAge)
		}
	}()

	log.Printf("relay listening on %s", cfg.Addr)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      httpapi.NewRouter(cfg, pool, registry, store, mailbox, hub, takeover, ptyHub, steerRequests),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("relay: listen on %s: %v", cfg.Addr, err)
	}
}
