package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

type permissionResolvePostRequest struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type permissionResolvePostResponse struct {
	Status   string `json:"status"`
	Decision string `json:"decision"`
}

func handlePermissionResolvePost(pool *db.Pool, store *stream.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")
		requestID := r.PathValue("request_id")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body permissionResolvePostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Decision != "allow" && body.Decision != "deny" {
			writeError(w, http.StatusBadRequest, `decision: required, must be "allow" or "deny"`)
			return
		}

		envelope, err := permissionResolvedEnvelope(sessionID, requestID, body.Decision, body.Reason, actor)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build permission resolution event")
			return
		}

		if _, err := publishEvent(r.Context(), pool, store, sessionID, envelope); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record permission resolution")
			return
		}

		writeJSON(w, http.StatusOK, permissionResolvePostResponse{Status: "resolved", Decision: body.Decision})
	}
}

func permissionResolvedEnvelope(sessionID, requestID, decision, reason string, actor auth.Actor) (json.RawMessage, error) {
	fields := map[string]any{
		"v":           1,
		"session_id":  sessionID,
		"ts":          time.Now().UTC().Format(time.RFC3339),
		"type":        "permission.resolved",
		"request_id":  requestID,
		"decision":    decision,
		"resolved_by": actorJSON(actor),
	}

	if reason != "" {
		fields["reason"] = map[string]any{"text": reason, "redactions": 0, "truncated": false}
	}

	return json.Marshal(fields)
}
