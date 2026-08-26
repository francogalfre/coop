package httpapi

import (
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/ratelimit"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
	"github.com/francogalfre/coop/apps/relay/internal/transport/wsapi"
)

func NewRouter(cfg config.Config, pool *db.Pool, registry *presence.Registry, store *stream.Store, mailbox *stream.Mailbox) http.Handler {
	mux := http.NewServeMux()

	ingestLimiter := ratelimit.New(ingestRatePerSecond, ingestBurst)
	steerLimiter := ratelimit.New(steerRatePerSecond, steerBurst)

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("POST /v1/events", withIPRateLimit(ingestLimiter, handleIngest(registry, store)))
	mux.HandleFunc("GET /v1/presence", handlePresence(registry))
	mux.HandleFunc("GET /v1/sessions", handleSessions(registry))
	mux.HandleFunc("GET /v1/sessions/{id}/stream", wsapi.NewSessionStreamHandler(store, cfg.WebOrigins))
	mux.HandleFunc("POST /v1/sessions/{id}/steer", withIPRateLimit(steerLimiter, handleSteerPost(mailbox)))
	mux.HandleFunc("GET /v1/sessions/{id}/steer", handleSteerGet(mailbox))
	mux.HandleFunc("POST /v1/auth/cli/exchange", handleCLIExchange(cfg, pool))

	return withCORS(mux, cfg.WebOrigins)
}
