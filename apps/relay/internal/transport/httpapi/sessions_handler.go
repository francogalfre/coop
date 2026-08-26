package httpapi

import (
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
)

type sessionsResponse struct {
	Repo     string                    `json:"repo"`
	Sessions []presence.SessionSummary `json:"sessions"`
}

func handleSessions(pool *db.Pool, registry *presence.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		if repo == "" {
			writeError(w, http.StatusBadRequest, "repo: required")
			return
		}

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		allowed, err := memberSessionIDs(r.Context(), pool, actor.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list projects")
			return
		}

		active := registry.ActiveSessions(repo)
		sessions := make([]presence.SessionSummary, 0, len(active))
		for _, sess := range active {
			if allowed[sess.SessionID] {
				sessions = append(sessions, sess)
			}
		}

		writeJSON(w, http.StatusOK, sessionsResponse{
			Repo:     repo,
			Sessions: sessions,
		})
	}
}
