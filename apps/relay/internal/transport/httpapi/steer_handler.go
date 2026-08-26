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

const steerTextMax = 4096

type steerPostRequest struct {
	Text string `json:"text"`
}

type steerGetResponse struct {
	HasMessage bool             `json:"has_message"`
	From       string           `json:"from,omitempty"`
	Text       string           `json:"text,omitempty"`
	Takeover   takeoverGetState `json:"takeover"`
}

type takeoverGetState struct {
	Active bool   `json:"active"`
	By     string `json:"by,omitempty"`
}

type steerPostResponse struct {
	Status string `json:"status"`
	Queued int    `json:"queued"`
}

func handleSteerPost(pool *db.Pool, mailbox *stream.Mailbox, store *stream.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body steerPostRequest

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Text == "" || len(body.Text) > steerTextMax {
			writeError(w, http.StatusBadRequest, "text: required, max length "+fmt.Sprint(steerTextMax))
			return
		}

		envelope, err := steerEnvelope(sessionID, actor, body.Text)
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

		mailbox.Put(sessionID, stream.SteerMessage{From: actor.DisplayName, Text: body.Text})

		writeJSON(w, http.StatusAccepted, steerPostResponse{
			Status: "accepted",
			Queued: mailbox.Depth(sessionID),
		})
	}
}

func steerEnvelope(sessionID string, actor auth.Actor, text string) (json.RawMessage, error) {
	fields := map[string]any{
		"v":          1,
		"session_id": sessionID,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "human.steer",
		"actor":      map[string]string{"id": actor.UserID, "display_name": actor.DisplayName},
		"text":       map[string]any{"text": text, "redactions": 0, "truncated": false},
	}

	return json.Marshal(fields)
}

func handleSteerGet(mailbox *stream.Mailbox, registry *stream.TakeoverRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		msg, hasMessage := mailbox.Take(sessionID)
		takeover := registry.Get(sessionID)

		resp := steerGetResponse{
			HasMessage: hasMessage,
			Takeover:   takeoverGetState{Active: takeover.Active, By: takeover.By},
		}
		if hasMessage {
			resp.From = msg.From
			resp.Text = msg.Text
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
