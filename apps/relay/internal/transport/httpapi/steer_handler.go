package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const steerTextMax = 4096

type steerMessageBody struct {
	From string `json:"from"`
	Text string `json:"text"`
}

func handleSteerPost(mailbox *stream.Mailbox) http.HandlerFunc {
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

		writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
	}
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
