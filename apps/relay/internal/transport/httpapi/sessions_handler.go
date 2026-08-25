package httpapi

import (
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/presence"
)

type sessionsResponse struct {
	Repo     string                    `json:"repo"`
	Sessions []presence.SessionSummary `json:"sessions"`
}

func handleSessions(registry *presence.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		if repo == "" {
			writeError(w, http.StatusBadRequest, "repo: required")
			return
		}

		writeJSON(w, http.StatusOK, sessionsResponse{
			Repo:     repo,
			Sessions: registry.ActiveSessions(repo),
		})
	}
}
