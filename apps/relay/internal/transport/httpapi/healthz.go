package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
)

const healthzTimeout = 5 * time.Second

func handleHealthz(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), healthzTimeout)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "unavailable",
				"error":  err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
