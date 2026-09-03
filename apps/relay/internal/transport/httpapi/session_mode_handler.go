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

type sessionModePostRequest struct {
	Mode string `json:"mode"`
}

type sessionModePostResponse struct {
	Mode string `json:"mode"`
}

func handleModePost(pool *db.Pool, store *stream.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body sessionModePostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Mode != db.SessionModeAuto && body.Mode != db.SessionModeRestricted {
			writeError(w, http.StatusBadRequest, `mode: required, must be "auto" or "restricted"`)
			return
		}

		if _, err := pool.SetSessionMode(r.Context(), sessionID, body.Mode); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update session mode")
			return
		}

		envelope, err := sessionModeChangedEnvelope(sessionID, body.Mode, actor)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build mode change event")
			return
		}

		dbEvent, err := pool.AppendEvent(r.Context(), sessionID, envelope)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record mode change")
			return
		}

		if _, err := store.AppendWithSeq(sessionID, dbEvent.Seq, envelope); err != nil {
			log.Printf("coop: failed to append mode change to in-memory store for session %s: %v", sessionID, err)
		}

		writeJSON(w, http.StatusOK, sessionModePostResponse{Mode: body.Mode})
	}
}

func sessionModeChangedEnvelope(sessionID, mode string, actor auth.Actor) (json.RawMessage, error) {
	fields := map[string]any{
		"v":          1,
		"session_id": sessionID,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "session.mode_changed",
		"mode":       mode,
		"changed_by": actorJSON(actor),
	}

	return json.Marshal(fields)
}
