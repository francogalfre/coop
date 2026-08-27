package httpapi

import (
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
)

func handleCLIRevoke(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		rawToken, ok := auth.BearerToken(r)
		if !ok {
			http.NotFound(w, r)
			return
		}

		if err := pool.RevokeCliCredential(r.Context(), actor.UserID, rawToken); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to revoke credential")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
	}
}
