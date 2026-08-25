package httpapi

import (
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
	"github.com/francogalfre/coop/apps/relay/internal/transport/wsapi"
)

func NewRouter(registry *presence.Registry, store *stream.Store, mailbox *stream.Mailbox) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("POST /v1/events", handleIngest(registry, store))
	mux.HandleFunc("GET /v1/presence", handlePresence(registry))
	mux.HandleFunc("GET /v1/sessions", handleSessions(registry))
	mux.HandleFunc("GET /v1/sessions/{id}/stream", wsapi.NewSessionStreamHandler(store))
	mux.HandleFunc("POST /v1/sessions/{id}/steer", handleSteerPost(mailbox))
	mux.HandleFunc("GET /v1/sessions/{id}/steer", handleSteerGet(mailbox))

	return mux
}
