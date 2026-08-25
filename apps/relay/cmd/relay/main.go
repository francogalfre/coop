package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
	"github.com/francogalfre/coop/apps/relay/internal/transport/httpapi"
)

const (
	sweepInterval = 5 * time.Minute
	sweepMaxAge   = 30 * time.Minute
)

func main() {
	cfg := config.Load()
	registry := presence.New()
	store := stream.New()
	mailbox := stream.NewMailbox()

	go func() {
		ticker := time.NewTicker(sweepInterval)
		defer ticker.Stop()

		for now := range ticker.C {
			registry.Sweep(now, sweepMaxAge)
		}
	}()

	log.Printf("relay listening on %s", cfg.Addr)

	if err := http.ListenAndServe(cfg.Addr, httpapi.NewRouter(registry, store, mailbox)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("relay: listen on %s: %v", cfg.Addr, err)
	}
}
