package httpapi

import (
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

type viewersResponse struct {
	Viewers []string `json:"viewers"`
}

func handleViewers(hub *stream.PresenceHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		writeJSON(w, http.StatusOK, viewersResponse{Viewers: hub.Viewers(sessionID)})
	}
}
