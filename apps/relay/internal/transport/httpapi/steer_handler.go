package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const steerTextMax = 4096

type steerMessageBody struct {
	From string `json:"from"`
	Text string `json:"text"`
}

type steerPostResponse struct {
	Status string `json:"status"`
	Queued int    `json:"queued"`
}

func handleSteerPost(mailbox *stream.Mailbox, store *stream.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body steerMessageBody

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.From == "" || len(body.From) > presence.DisplayNameMax {
			writeError(w, http.StatusBadRequest, "from: required, max length "+fmt.Sprint(presence.DisplayNameMax))
			return
		}

		if body.Text == "" || len(body.Text) > steerTextMax {
			writeError(w, http.StatusBadRequest, "text: required, max length "+fmt.Sprint(steerTextMax))
			return
		}

		mailbox.Put(sessionID, stream.SteerMessage{From: body.From, Text: body.Text})

		if envelope, err := steerEnvelope(sessionID, body); err == nil {
			_, _ = store.Append(sessionID, envelope)
		}

		writeJSON(w, http.StatusAccepted, steerPostResponse{
			Status: "accepted",
			Queued: mailbox.Depth(sessionID),
		})
	}
}

func steerEnvelope(sessionID string, body steerMessageBody) (json.RawMessage, error) {
	fields := map[string]any{
		"v":          1,
		"session_id": sessionID,
		"seq":        0,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "human.steer",
		"actor":      map[string]string{"id": body.From, "display_name": body.From},
		"text":       map[string]any{"text": body.Text, "redactions": 0, "truncated": false},
	}

	return json.Marshal(fields)
}

func handleSteerGet(mailbox *stream.Mailbox) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		msg, ok := mailbox.Take(sessionID)
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		writeJSON(w, http.StatusOK, steerMessageBody{From: msg.From, Text: msg.Text})
	}
}
