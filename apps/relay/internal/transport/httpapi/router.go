package httpapi

import (
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/presence"
)

func NewRouter(registry *presence.Registry) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("POST /v1/events", handleIngest(registry))
	mux.HandleFunc("GET /v1/presence", handlePresence(registry))
	mux.HandleFunc("GET /v1/sessions", handleSessions(registry))

	return mux
}
