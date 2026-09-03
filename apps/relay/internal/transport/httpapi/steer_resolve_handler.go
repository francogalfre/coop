package httpapi

import (
	"encoding/json"
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

		pending, ok, err := steerRequests.Take(r.Context(), sessionID, requestID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to look up pending steer request")
			return
		}
		if !ok {
			http.NotFound(w, r)
			return
		}

		if body.Decision == "allow" {
			steerID, err := randomID()
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to build steer message")
				return
			}

			envelope, err := steerEnvelope(sessionID, pending.Actor, pending.Text, requestID, steerID, "", nil)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to build steer message")
				return
			}

			if _, err := publishEvent(r.Context(), pool, store, sessionID, envelope); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to record steer message")
				return
			}

			dropped, wasDropped := mailbox.Put(sessionID, stream.SteerMessage{ID: steerID, Kind: "steer", From: pending.Actor.DisplayName, Text: pending.Text})
			if wasDropped {
				emitSteerDropped(r.Context(), pool, store, sessionID, dropped.ID, "queue_overflow")
			}
		}

		resolvedEnvelope, err := steerResolvedEnvelope(sessionID, requestID, body.Decision, actor)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build steer resolution event")
			return
		}

		if _, err := publishEvent(r.Context(), pool, store, sessionID, resolvedEnvelope); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record steer resolution")
			return
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
		"resolved_by": actorJSON(actor),
	}

	return json.Marshal(fields)
}
