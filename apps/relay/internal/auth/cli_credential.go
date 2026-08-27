package auth

import (
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/francogalfre/coop/apps/relay/internal/db"
)

const bearerPrefix = "Bearer "

func BearerToken(r *http.Request) ([]byte, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, bearerPrefix) {
		return nil, false
	}

	rawToken, err := hex.DecodeString(strings.TrimPrefix(header, bearerPrefix))
	if err != nil {
		return nil, false
	}

	return rawToken, true
}

func RequireCliCredential(pool *db.Pool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rawToken, ok := BearerToken(r)
			if !ok {
				http.NotFound(w, r)
				return
			}

			userID, displayName, err := pool.AuthenticateCliCredential(r.Context(), rawToken)
			if err != nil {
				if db.IsNotFound(err) {
					http.NotFound(w, r)
					return
				}

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			next(w, r.WithContext(WithActor(r.Context(), Actor{UserID: userID, DisplayName: displayName})))
		}
	}
}
