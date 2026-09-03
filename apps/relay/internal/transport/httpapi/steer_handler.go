package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const (
	steerTextMax = 4096
)

type steerPostRequest struct {
	Text     string `json:"text"`
	ClientID string `json:"client_id"`
}

type steerGetResponse struct {
	HasMessage bool             `json:"has_message"`
	ID         string           `json:"id,omitempty"`
	Kind       string           `json:"kind,omitempty"`
	From       string           `json:"from,omitempty"`
	Text       string           `json:"text,omitempty"`
	Takeover   takeoverGetState `json:"takeover"`
}

type takeoverGetState struct {
	Active bool   `json:"active"`
	By     string `json:"by,omitempty"`
}

type steerPostResponse struct {
	Status    string `json:"status"`
	Queued    int    `json:"queued,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func handleSteerPost(pool *db.Pool, mailbox *stream.Mailbox, store *stream.Store, steerRequests *stream.SteerRequestRegistry) http.HandlerFunc {
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

		ownerID, mode, err := steerSessionState(r.Context(), pool, sessionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to look up agent session")
			return
		}

		if mode == db.SessionModeRestricted && actor.UserID != ownerID {
			handleSteerRequestPending(w, r, pool, store, steerRequests, sessionID, actor, body.Text)
			return
		}

		deliverSteerNow(w, r, pool, mailbox, store, sessionID, actor, body.Text, body.ClientID)
	}
}

func deliverSteerNow(w http.ResponseWriter, r *http.Request, pool *db.Pool, mailbox *stream.Mailbox, store *stream.Store, sessionID string, actor auth.Actor, text, clientID string) {
	steerID, err := randomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build steer message")
		return
	}

	envelope, err := steerEnvelope(sessionID, actor, text, "", steerID, clientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build steer message")
		return
	}

	if _, err := publishEvent(r.Context(), pool, store, sessionID, envelope); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record steer message")
		return
	}

	dropped, wasDropped := mailbox.Put(sessionID, stream.SteerMessage{ID: steerID, Kind: "steer", From: actor.DisplayName, Text: text})
	if wasDropped {
		emitSteerDropped(r.Context(), pool, store, sessionID, dropped.ID, "queue_overflow")
	}

	writeJSON(w, http.StatusAccepted, steerPostResponse{
		Status: "accepted",
		Queued: mailbox.Depth(sessionID),
	})
}

func emitSteerDropped(ctx context.Context, pool *db.Pool, store *stream.Store, sessionID, steerID, reason string) {
	if steerID == "" {
		return
	}

	envelope, err := json.Marshal(map[string]any{
		"v":          1,
		"session_id": sessionID,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "steer.dropped",
		"steer_id":   steerID,
		"reason":     reason,
	})
	if err != nil {
		log.Printf("coop: failed to build steer.dropped event for session %s: %v", sessionID, err)
		return
	}

	if _, err := publishEvent(ctx, pool, store, sessionID, envelope); err != nil {
		log.Printf("coop: failed to record steer.dropped event for session %s: %v", sessionID, err)
	}
}

func handleSteerRequestPending(w http.ResponseWriter, r *http.Request, pool *db.Pool, store *stream.Store, steerRequests *stream.SteerRequestRegistry, sessionID string, actor auth.Actor, text string) {
	requestID, err := randomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build steer request")
		return
	}

	envelope, err := steerRequestedEnvelope(sessionID, requestID, actor, text)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build steer request event")
		return
	}

	if _, err := publishEvent(r.Context(), pool, store, sessionID, envelope); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record steer request")
		return
	}

	if err := steerRequests.Put(r.Context(), sessionID, stream.PendingSteerRequest{RequestID: requestID, Actor: actor, Text: text}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record pending steer request")
		return
	}

	writeJSON(w, http.StatusAccepted, steerPostResponse{
		Status:    "pending",
		RequestID: requestID,
	})
}

func steerSessionState(ctx context.Context, pool *db.Pool, sessionID string) (ownerID, mode string, err error) {
	if pool == nil {
		return "", db.SessionModeAuto, nil
	}

	sess, err := pool.GetAgentSession(ctx, sessionID)
	if err != nil {
		if db.IsNotFound(err) {
			return "", db.SessionModeAuto, nil
		}

		return "", "", err
	}

	return sess.OwnerID, sess.Mode, nil
}

func steerEnvelope(sessionID string, actor auth.Actor, text, requestID, steerID, clientID string) (json.RawMessage, error) {
	fields := map[string]any{
		"v":          1,
		"session_id": sessionID,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "human.steer",
		"actor":      actorJSON(actor),
		"text":       map[string]any{"text": text, "redactions": 0, "truncated": false},
	}

	// request_id lets the web dedupe this against the steer.requested bubble it approved.
	if requestID != "" {
		fields["request_id"] = requestID
	}

	// steer_id ties this message to its later steer.delivered/steer.dropped receipt.
	if steerID != "" {
		fields["steer_id"] = steerID
	}

	// client_id lets the composer reconcile its optimistic echo with this real event.
	if clientID != "" {
		fields["client_id"] = clientID
	}

	return json.Marshal(fields)
}

func steerRequestedEnvelope(sessionID, requestID string, actor auth.Actor, text string) (json.RawMessage, error) {
	fields := map[string]any{
		"v":          1,
		"session_id": sessionID,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "steer.requested",
		"request_id": requestID,
		"actor":      actorJSON(actor),
		"text":       map[string]any{"text": text, "redactions": 0, "truncated": false},
	}

	return json.Marshal(fields)
}

func handleSteerGet(mailbox *stream.Mailbox, registry *stream.TakeoverRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		msg, hasMessage := mailbox.Take(sessionID)
		takeover, err := registry.Get(r.Context(), sessionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to look up takeover state")
			return
		}

		resp := steerGetResponse{
			HasMessage: hasMessage,
			Takeover:   takeoverGetState{Active: takeover.Active, By: takeover.By},
		}
		if hasMessage {
			resp.ID = msg.ID
			resp.Kind = msg.Kind
			resp.From = msg.From
			resp.Text = msg.Text
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
