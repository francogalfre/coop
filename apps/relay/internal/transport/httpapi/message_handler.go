package httpapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const messageTextMax = 4096

type messagePostRequest struct {
	Text      string `json:"text"`
	AnchorSeq *int   `json:"anchor_seq"`
}

type messagePostResponse struct {
	Status string `json:"status"`
	Seq    int    `json:"seq"`
}

func handleMessagePost(pool *db.Pool, store *stream.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body messagePostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Text == "" || len(body.Text) > messageTextMax {
			writeError(w, http.StatusBadRequest, "text: required, max length "+fmt.Sprint(messageTextMax))
			return
		}

		if body.AnchorSeq != nil && *body.AnchorSeq < 0 {
			writeError(w, http.StatusBadRequest, "anchor_seq: must be a non-negative integer")
			return
		}

		envelope, err := messageEnvelope(sessionID, actor, body.Text, body.AnchorSeq)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build message")
			return
		}

		dbEvent, err := pool.AppendEvent(r.Context(), sessionID, envelope)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record message")
			return
		}

		if _, err := store.AppendWithSeq(sessionID, dbEvent.Seq, envelope); err != nil {
			log.Printf("coop: failed to append message to in-memory store for session %s: %v", sessionID, err)
		}

		writeJSON(w, http.StatusAccepted, messagePostResponse{Status: "sent", Seq: dbEvent.Seq})
	}
}

func messageEnvelope(sessionID string, actor auth.Actor, text string, anchorSeq *int) (json.RawMessage, error) {
	fields := map[string]any{
		"v":          1,
		"session_id": sessionID,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "human.message",
		"actor":      map[string]string{"id": actor.UserID, "display_name": actor.DisplayName},
		"text":       map[string]any{"text": text, "redactions": 0, "truncated": false},
	}

	if anchorSeq != nil {
		fields["anchor_seq"] = *anchorSeq
	}

	return json.Marshal(fields)
}
