package httpapi

import (
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/ratelimit"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
	"github.com/francogalfre/coop/apps/relay/internal/transport/wsapi"
)

func NewRouter(cfg config.Config, pool *db.Pool, registry *presence.Registry, store *stream.Store, mailbox *stream.Mailbox, hub *stream.PresenceHub) http.Handler {
	mux := http.NewServeMux()

	ingestLimiter := ratelimit.New(ingestRatePerSecond, ingestBurst)
	steerLimiter := ratelimit.New(steerRatePerSecond, steerBurst)

	requireIdentity := auth.RequireAnyIdentity(pool, cfg)
	optionalIdentity := auth.OptionalCliCredential(pool)

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("POST /v1/events", withIPRateLimit(ingestLimiter, optionalIdentity(handleIngest(pool, registry, store))))
	mux.HandleFunc("GET /v1/presence", handlePresence(registry))
	mux.HandleFunc("GET /v1/sessions", handleSessions(registry))
	mux.HandleFunc("GET /v1/sessions/{id}/stream", wsapi.NewSessionStreamHandler(store, hub, cfg.WebOrigins))
	mux.HandleFunc("GET /v1/sessions/{id}/viewers", handleViewers(hub))
	mux.HandleFunc("POST /v1/sessions/{id}/steer", withIPRateLimit(steerLimiter, handleSteerPost(mailbox, store)))
	mux.HandleFunc("GET /v1/sessions/{id}/steer", handleSteerGet(mailbox))
	mux.HandleFunc("POST /v1/auth/cli/exchange", handleCLIExchange(cfg, pool))

	mux.HandleFunc("POST /v1/projects", requireIdentity(handleCreateProject(pool)))
	mux.HandleFunc("GET /v1/projects", requireIdentity(handleListUserProjects(pool)))
	mux.HandleFunc("POST /v1/projects/{slug}/invites", requireIdentity(handleCreateInvite(pool)))
	mux.HandleFunc("POST /v1/projects/invites/{token}/accept", requireIdentity(handleAcceptInvite(pool)))
	mux.HandleFunc("GET /v1/projects/{slug}/sessions", requireIdentity(handleListProjectSessions(pool)))

	return withCORS(mux, cfg.WebOrigins)
}
