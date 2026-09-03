package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
	"github.com/francogalfre/coop/apps/relay/internal/transport/httpapi"
)

const (
	sweepInterval   = 5 * time.Minute
	sweepMaxAge     = 30 * time.Minute
	shutdownTimeout = 10 * time.Second
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("relay: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	takeover := stream.NewTakeoverRegistry(pool)
	ptyHub := stream.NewPtyHub()
	steerRequests := stream.NewSteerRequestRegistry(pool)
	questions := stream.NewQuestionRegistry()

	go func() {
		ticker := time.NewTicker(sweepInterval)
		defer ticker.Stop()

		for now := range ticker.C {
			registry.Sweep(now, sweepMaxAge)
		}
	}()

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      httpapi.NewRouter(cfg, pool, registry, store, mailbox, hub, takeover, ptyHub, steerRequests, questions),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("relay listening on %s", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			log.Fatalf("relay: listen on %s: %v", cfg.Addr, err)
		}
	case <-ctx.Done():
		stop()
		log.Print("relay: shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("relay: shutdown: %v", err)
		}
	}
}
