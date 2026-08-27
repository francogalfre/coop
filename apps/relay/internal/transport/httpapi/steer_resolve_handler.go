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

type steerResolvePostRequest struct {
	Decision string `json:"decision"`
}

type steerResolvePostResponse struct {
	Status   string `json:"status"`
	Decision string `json:"decision"`
}

func handleSteerResolvePost(pool *db.Pool, mailbox *stream.Mailbox, store *stream.Store, steerRequests *stream.SteerRequestRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")
		requestID := r.PathValue("request_id")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body steerResolvePostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Decision != "allow" && body.Decision != "deny" {
			writeError(w, http.StatusBadRequest, `decision: required, must be "allow" or "deny"`)
			return
		}

		pending, ok := steerRequests.Take(sessionID, requestID)
		if !ok {
			http.NotFound(w, r)
			return
		}

		if body.Decision == "allow" {
			envelope, err := steerEnvelope(sessionID, pending.Actor, pending.Text, requestID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to build steer message")
				return
			}

			dbEvent, err := pool.AppendEvent(r.Context(), sessionID, envelope)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to record steer message")
				return
			}

			if _, err := store.AppendWithSeq(sessionID, dbEvent.Seq, envelope); err != nil {
				log.Printf("coop: failed to append steer message to in-memory store for session %s: %v", sessionID, err)
			}

			mailbox.Put(sessionID, stream.SteerMessage{From: pending.Actor.DisplayName, Text: pending.Text})
		}

		resolvedEnvelope, err := steerResolvedEnvelope(sessionID, requestID, body.Decision, actor)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build steer resolution event")
			return
		}

		dbEvent, err := pool.AppendEvent(r.Context(), sessionID, resolvedEnvelope)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record steer resolution")
			return
		}

		if _, err := store.AppendWithSeq(sessionID, dbEvent.Seq, resolvedEnvelope); err != nil {
			log.Printf("coop: failed to append steer resolution to in-memory store for session %s: %v", sessionID, err)
		}

		writeJSON(w, http.StatusOK, steerResolvePostResponse{Status: "resolved", Decision: body.Decision})
	}
}

func steerResolvedEnvelope(sessionID, requestID, decision string, actor auth.Actor) (json.RawMessage, error) {
	fields := map[string]any{
		"v":           1,
		"session_id":  sessionID,
		"ts":          time.Now().UTC().Format(time.RFC3339),
		"type":        "steer.resolved",
		"request_id":  requestID,
		"decision":    decision,
		"resolved_by": map[string]string{"id": actor.UserID, "display_name": actor.DisplayName},
	}

	return json.Marshal(fields)
}
