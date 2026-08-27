package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func restrictedSessionFixture(t *testing.T, sessionID, ownerID string) *db.Pool {
	t.Helper()

	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", ownerID)
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := pool.CreateAgentSession(t.Context(), sessionID, proj, ownerID, "/repo", "/repo", "claude-code", time.Now()); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	if _, err := pool.SetSessionMode(t.Context(), sessionID, db.SessionModeRestricted); err != nil {
		t.Fatalf("SetSessionMode: %v", err)
	}

	return pool
}

func TestSteerPostRestrictedModeHoldsNonOwnerMessage(t *testing.T) {
	pool := restrictedSessionFixture(t, "sess-a", "user-owner")
	mailbox := stream.NewMailbox()
	store := stream.New()
	steerRequests := stream.NewSteerRequestRegistry()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/steer", strings.NewReader(`{"text":"try the other branch"}`))
	req.SetPathValue("id", "sess-a")
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handleSteerPost(pool, mailbox, store, steerRequests)(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	var payload steerPostResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.Status != "pending" || payload.RequestID == "" {
		t.Fatalf("got payload %+v, want pending status with a request id", payload)
	}

	if _, hasMessage := mailbox.Take("sess-a"); hasMessage {
		t.Fatal("got message in mailbox, want it held pending owner approval")
	}

	events := store.Since("sess-a", 0)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	var fields map[string]any
	if err := json.Unmarshal(events[0].Data, &fields); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if fields["type"] != "steer.requested" {
		t.Fatalf("got type %v, want steer.requested", fields["type"])
	}
	if fields["request_id"] != payload.RequestID {
		t.Fatalf("got request_id %v, want %v", fields["request_id"], payload.RequestID)
	}
}

func TestSteerPostRestrictedModeOwnerBypasses(t *testing.T) {
	pool := restrictedSessionFixture(t, "sess-a", "user-owner")
	mailbox := stream.NewMailbox()
	store := stream.New()
	steerRequests := stream.NewSteerRequestRegistry()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/steer", strings.NewReader(`{"text":"try the other branch"}`))
	req.SetPathValue("id", "sess-a")
	req = withActorNamed(req, "user-owner", "Owner")
	rec := httptest.NewRecorder()

	handleSteerPost(pool, mailbox, store, steerRequests)(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	var payload steerPostResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.Status != "accepted" {
		t.Fatalf("got status %q, want accepted (owner bypasses restriction)", payload.Status)
	}

	if _, hasMessage := mailbox.Take("sess-a"); !hasMessage {
		t.Fatal("got no message in mailbox, want immediate delivery for the owner")
	}
}
