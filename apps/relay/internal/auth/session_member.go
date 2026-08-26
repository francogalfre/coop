package auth

import (
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
)

func RequireSessionMember(pool *db.Pool, cfg config.Config) func(http.HandlerFunc) http.HandlerFunc {
	requireIdentity := RequireAnyIdentity(pool, cfg)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return requireIdentity(func(w http.ResponseWriter, r *http.Request) {
			if !isSessionMember(r, pool) {
				http.NotFound(w, r)
				return
			}

			next(w, r)
		})
	}
}

func RequireSessionOwner(pool *db.Pool, cfg config.Config) func(http.HandlerFunc) http.HandlerFunc {
	requireIdentity := RequireAnyIdentity(pool, cfg)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return requireIdentity(func(w http.ResponseWriter, r *http.Request) {
			if !isSessionOwner(r, pool) {
				http.NotFound(w, r)
				return
			}

			next(w, r)
		})
	}
}

func isSessionMember(r *http.Request, pool *db.Pool) bool {
	actor, ok := FromContext(r.Context())
	if !ok {
		return false
	}

	sess, err := pool.GetAgentSession(r.Context(), r.PathValue("id"))
	if err != nil || sess.Edges.Project == nil {
		return false
	}

	_, isMember, err := pool.MemberRole(r.Context(), sess.Edges.Project.ID, actor.UserID)
	return err == nil && isMember
}

func isSessionOwner(r *http.Request, pool *db.Pool) bool {
	actor, ok := FromContext(r.Context())
	if !ok {
		return false
	}

	sess, err := pool.GetAgentSession(r.Context(), r.PathValue("id"))
	if err != nil {
		return false
	}

	return sess.OwnerID == actor.UserID
}
