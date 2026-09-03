package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const commandArgsMax = 256

type commandPostRequest struct {
	Command string `json:"command"`
	Args    string `json:"args"`
}

type commandPostResponse struct {
	Status string `json:"status"`
	Queued int    `json:"queued,omitempty"`
}

// handleCommandPost delivers a harness command (e.g. /model) into the session owner's own terminal.
func handleCommandPost(pool *db.Pool, mailbox *stream.Mailbox, store *stream.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body commandPostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if !harnessCommands[body.Command] {
			writeError(w, http.StatusBadRequest, "command: not on the allowlist")
			return
		}

		if len(body.Args) > commandArgsMax {
			writeError(w, http.StatusBadRequest, "args: max length "+fmt.Sprint(commandArgsMax))
			return
		}

		envelope, err := commandEnvelope(sessionID, actor, body.Command, body.Args)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build command event")
			return
		}

		if _, err := publishEvent(r.Context(), pool, store, sessionID, envelope); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record command")
			return
		}

		commandID, err := randomID()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build command")
			return
		}

		keystrokes := "/" + body.Command
		if body.Args != "" {
			keystrokes += " " + body.Args
		}

		dropped, wasDropped := mailbox.Put(sessionID, stream.SteerMessage{ID: commandID, Kind: "command", From: actor.DisplayName, Text: keystrokes})
		if wasDropped {
			emitSteerDropped(r.Context(), pool, store, sessionID, dropped.ID, "queue_overflow")
		}

		writeJSON(w, http.StatusAccepted, commandPostResponse{Status: "accepted", Queued: mailbox.Depth(sessionID)})
	}
}

func commandEnvelope(sessionID string, actor auth.Actor, command, args string) (json.RawMessage, error) {
	fields := map[string]any{
		"v":          1,
		"session_id": sessionID,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "human.command",
		"actor":      actorJSON(actor),
		"command":    command,
	}

	if strings.TrimSpace(args) != "" {
		fields["args"] = args
	}

	return json.Marshal(fields)
}
