package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func TestHandleCommandPostDeliversAnAllowlistedCommand(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	mailbox := stream.NewMailbox()
	store := stream.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/command", strings.NewReader(`{"command":"model","args":"sonnet"}`))
	req.SetPathValue("id", "sess-a")
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handleCommandPost(pool, mailbox, store)(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	msg, ok := mailbox.Take("sess-a")
	if !ok {
		t.Fatal("got no mailbox message, want the command queued for delivery")
	}
	if msg.Kind != "command" || msg.Text != "/model sonnet" {
		t.Fatalf("got %+v, want kind=command text=\"/model sonnet\"", msg)
	}

	events := store.Since("sess-a", 0)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	var event struct {
		Type    string `json:"type"`
		Command string `json:"command"`
		Args    string `json:"args"`
	}
	if err := json.Unmarshal(events[0].Data, &event); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if event.Type != "human.command" || event.Command != "model" || event.Args != "sonnet" {
		t.Fatalf("got %+v, want human.command model sonnet", event)
	}
}

func TestHandleCommandPostRejectsCommandOffAllowlist(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	mailbox := stream.NewMailbox()
	store := stream.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/command", strings.NewReader(`{"command":"bash","args":"rm -rf /"}`))
	req.SetPathValue("id", "sess-a")
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handleCommandPost(pool, mailbox, store)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
	if _, hasMessage := mailbox.Take("sess-a"); hasMessage {
		t.Fatal("got a queued message for a rejected command, want none")
	}
}

func TestHandleCommandPostRejectsShellFragmentSmuggledAsCommand(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	mailbox := stream.NewMailbox()
	store := stream.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/command", strings.NewReader(`{"command":"model; rm -rf /"}`))
	req.SetPathValue("id", "sess-a")
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handleCommandPost(pool, mailbox, store)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}
