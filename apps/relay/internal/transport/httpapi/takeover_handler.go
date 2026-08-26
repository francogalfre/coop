package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

type takeoverPostRequest struct {
	Active bool `json:"active"`
}

type takeoverResponse struct {
	Active bool   `json:"active"`
	By     string `json:"by,omitempty"`
}

func handleTakeoverPost(pool *db.Pool, store *stream.Store, registry *stream.TakeoverRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body takeoverPostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		current := registry.Get(sessionID)

		if body.Active && current.Active && current.ByID != actor.UserID {
			writeJSON(w, http.StatusConflict, takeoverResponse{Active: true, By: current.By})
			return
		}

		if !body.Active && current.Active && current.ByID != actor.UserID {
			sess, err := pool.GetAgentSession(r.Context(), sessionID)
			if err != nil || sess.OwnerID != actor.UserID {
				writeJSON(w, http.StatusForbidden, takeoverResponse{Active: true, By: current.By})
				return
			}
		}

		envelope, err := takeoverEnvelope(sessionID, actor, body.Active)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build takeover event")
			return
		}

		dbEvent, err := pool.AppendEvent(r.Context(), sessionID, envelope)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record takeover event")
			return
		}

		if _, err := store.AppendWithSeq(sessionID, dbEvent.Seq, envelope); err != nil {
			log.Printf("coop: failed to append takeover event to in-memory store for session %s: %v", sessionID, err)
		}

		registry.Set(sessionID, stream.TakeoverState{Active: body.Active, ByID: actor.UserID, By: actor.DisplayName})

		writeJSON(w, http.StatusOK, takeoverResponse{Active: body.Active, By: actor.DisplayName})
	}
}

func takeoverEnvelope(sessionID string, actor auth.Actor, active bool) (json.RawMessage, error) {
	fields := map[string]any{
		"v":          1,
		"session_id": sessionID,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "human.takeover",
		"actor":      map[string]string{"id": actor.UserID, "display_name": actor.DisplayName},
		"active":     active,
	}

	return json.Marshal(fields)
}
