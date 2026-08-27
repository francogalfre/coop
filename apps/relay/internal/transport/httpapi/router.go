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

func NewRouter(cfg config.Config, pool *db.Pool, registry *presence.Registry, store *stream.Store, mailbox *stream.Mailbox, hub *stream.PresenceHub, takeover *stream.TakeoverRegistry, ptyHub *stream.PtyHub, steerRequests *stream.SteerRequestRegistry) http.Handler {
	mux := http.NewServeMux()

	ingestLimiter := ratelimit.New(ingestRatePerSecond, ingestBurst)
	steerLimiter := ratelimit.New(steerRatePerSecond, steerBurst)
	exchangeLimiter := ratelimit.New(exchangeRatePerSecond, exchangeBurst)

	requireIdentity := auth.RequireAnyIdentity(pool, cfg)
	requireCliCredential := auth.RequireCliCredential(pool)
	requireSessionMember := auth.RequireSessionMember(pool, cfg)
	requireSessionOwner := auth.RequireSessionOwner(pool, cfg)

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("POST /v1/events", withIPRateLimit(ingestLimiter, requireCliCredential(handleIngest(pool, registry, store))))
	mux.HandleFunc("GET /v1/presence", requireIdentity(handlePresence(pool, registry)))
	mux.HandleFunc("GET /v1/sessions", requireIdentity(handleSessions(pool, registry)))
	mux.HandleFunc("GET /v1/sessions/{id}/stream", requireSessionMember(wsapi.NewSessionStreamHandler(pool, store, hub, cfg.WebOrigins)))
	mux.HandleFunc("GET /v1/sessions/{id}/pty", requireSessionMember(wsapi.NewPtySessionHandler(ptyHub, takeover, pool, cfg.WebOrigins)))
	mux.HandleFunc("GET /v1/sessions/{id}/viewers", requireSessionMember(handleViewers(hub)))
	mux.HandleFunc("POST /v1/sessions/{id}/steer", withIPRateLimit(steerLimiter, requireSessionMember(handleSteerPost(pool, mailbox, store, steerRequests))))
	mux.HandleFunc("GET /v1/sessions/{id}/steer", requireSessionOwner(handleSteerGet(mailbox, takeover)))
	mux.HandleFunc("POST /v1/sessions/{id}/steer/{request_id}/resolve", requireSessionOwner(handleSteerResolvePost(pool, mailbox, store, steerRequests)))
	mux.HandleFunc("GET /v1/sessions/{id}/events", requireSessionMember(handleEvents(pool)))
	mux.HandleFunc("POST /v1/sessions/{id}/takeover", requireSessionMember(handleTakeoverPost(pool, store, takeover)))
	mux.HandleFunc("POST /v1/sessions/{id}/mode", requireSessionOwner(handleModePost(pool, store)))
	mux.HandleFunc("POST /v1/auth/cli/exchange", withIPRateLimit(exchangeLimiter, handleCLIExchange(cfg, pool)))
	mux.HandleFunc("POST /v1/auth/cli/revoke", requireCliCredential(handleCLIRevoke(pool)))

	mux.HandleFunc("POST /v1/projects", requireIdentity(handleCreateProject(pool)))
	mux.HandleFunc("GET /v1/projects", requireIdentity(handleListUserProjects(pool)))
	mux.HandleFunc("POST /v1/projects/{slug}/invites", requireIdentity(handleCreateInvite(pool)))
	mux.HandleFunc("POST /v1/projects/invites/{token}/accept", requireIdentity(handleAcceptInvite(pool)))
	mux.HandleFunc("GET /v1/projects/{slug}/sessions", requireIdentity(handleListProjectSessions(pool)))

	return withCORS(mux, cfg.WebOrigins)
}
