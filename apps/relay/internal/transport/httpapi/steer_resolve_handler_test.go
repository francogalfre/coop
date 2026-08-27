package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func requestPendingSteer(t *testing.T, pool *db.Pool, store *stream.Store, steerRequests *stream.SteerRequestRegistry, sessionID string) string {
	t.Helper()

	mailbox := stream.NewMailbox()
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/steer", strings.NewReader(`{"text":"try the other branch"}`))
	req.SetPathValue("id", sessionID)
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handleSteerPost(pool, mailbox, store, steerRequests)(rec, req)

	var payload steerPostResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode pending steer response: %v", err)
	}
	if payload.RequestID == "" {
		t.Fatalf("got empty request id from pending steer post: %s", rec.Body.String())
	}

	return payload.RequestID
}

func doSteerResolvePost(t *testing.T, pool *db.Pool, mailbox *stream.Mailbox, store *stream.Store, steerRequests *stream.SteerRequestRegistry, sessionID, requestID, userID, displayName, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/steer/"+requestID+"/resolve", strings.NewReader(body))
	req.SetPathValue("id", sessionID)
	req.SetPathValue("request_id", requestID)
	req = withActorNamed(req, userID, displayName)
	rec := httptest.NewRecorder()

	handleSteerResolvePost(pool, mailbox, store, steerRequests)(rec, req)

	return rec
}

func TestSteerResolveAllowDeliversToMailbox(t *testing.T) {
	pool := restrictedSessionFixture(t, "sess-a", "user-owner")
	mailbox := stream.NewMailbox()
	store := stream.New()
	steerRequests := stream.NewSteerRequestRegistry()

	requestID := requestPendingSteer(t, pool, store, steerRequests, "sess-a")

	rec := doSteerResolvePost(t, pool, mailbox, store, steerRequests, "sess-a", requestID, "user-owner", "Owner", `{"decision":"allow"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	msg, hasMessage := mailbox.Take("sess-a")
	if !hasMessage || msg.Text != "try the other branch" || msg.From != "Alice" {
		t.Fatalf("got mailbox message %+v (hasMessage=%v), want Alice's message delivered", msg, hasMessage)
	}

	events := store.Since("sess-a", 0)
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3 (steer.requested + human.steer + steer.resolved)", len(events))
	}

	var delivered map[string]any
	if err := json.Unmarshal(events[1].Data, &delivered); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if delivered["type"] != "human.steer" {
		t.Fatalf("got event %+v, want human.steer delivered on allow", delivered)
	}

	var resolved map[string]any
	if err := json.Unmarshal(events[2].Data, &resolved); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if resolved["type"] != "steer.resolved" || resolved["decision"] != "allow" {
		t.Fatalf("got event %+v, want steer.resolved with decision allow", resolved)
	}
}

func TestSteerResolveDenyLeavesMailboxEmpty(t *testing.T) {
	pool := restrictedSessionFixture(t, "sess-a", "user-owner")
	mailbox := stream.NewMailbox()
	store := stream.New()
	steerRequests := stream.NewSteerRequestRegistry()

	requestID := requestPendingSteer(t, pool, store, steerRequests, "sess-a")

	rec := doSteerResolvePost(t, pool, mailbox, store, steerRequests, "sess-a", requestID, "user-owner", "Owner", `{"decision":"deny"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	if _, hasMessage := mailbox.Take("sess-a"); hasMessage {
		t.Fatal("got message in mailbox, want a denied message to never reach it")
	}

	events := store.Since("sess-a", 0)
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2 (steer.requested + steer.resolved)", len(events))
	}

	var resolved map[string]any
	if err := json.Unmarshal(events[1].Data, &resolved); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if resolved["type"] != "steer.resolved" || resolved["decision"] != "deny" {
		t.Fatalf("got event %+v, want steer.resolved with decision deny", resolved)
	}
}

func TestSteerResolveUnknownRequestIDNotFound(t *testing.T) {
	pool := restrictedSessionFixture(t, "sess-a", "user-owner")
	mailbox := stream.NewMailbox()
	store := stream.New()
	steerRequests := stream.NewSteerRequestRegistry()

	rec := doSteerResolvePost(t, pool, mailbox, store, steerRequests, "sess-a", "bogus-request-id", "user-owner", "Owner", `{"decision":"allow"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestSteerResolveAlreadyResolvedNotFound(t *testing.T) {
	pool := restrictedSessionFixture(t, "sess-a", "user-owner")
	mailbox := stream.NewMailbox()
	store := stream.New()
	steerRequests := stream.NewSteerRequestRegistry()

	requestID := requestPendingSteer(t, pool, store, steerRequests, "sess-a")

	doSteerResolvePost(t, pool, mailbox, store, steerRequests, "sess-a", requestID, "user-owner", "Owner", `{"decision":"allow"}`)

	rec := doSteerResolvePost(t, pool, mailbox, store, steerRequests, "sess-a", requestID, "user-owner", "Owner", `{"decision":"allow"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404 for already-resolved request: %s", rec.Code, rec.Body.String())
	}
}

func TestSteerResolveInvalidDecisionRejected(t *testing.T) {
	pool := restrictedSessionFixture(t, "sess-a", "user-owner")
	mailbox := stream.NewMailbox()
	store := stream.New()
	steerRequests := stream.NewSteerRequestRegistry()

	requestID := requestPendingSteer(t, pool, store, steerRequests, "sess-a")

	rec := doSteerResolvePost(t, pool, mailbox, store, steerRequests, "sess-a", requestID, "user-owner", "Owner", `{"decision":"maybe"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func TestSteerResolveRouteRejectsNonOwner(t *testing.T) {
	pool := restrictedSessionFixture(t, "sess-a", "user-owner")
	mailbox := stream.NewMailbox()
	store := stream.New()
	steerRequests := stream.NewSteerRequestRegistry()
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	requestID := requestPendingSteer(t, pool, store, steerRequests, "sess-a")

	handler := auth.RequireSessionOwner(pool, cfg)(handleSteerResolvePost(pool, mailbox, store, steerRequests))

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/steer/"+requestID+"/resolve", strings.NewReader(`{"decision":"allow"}`))
	req.SetPathValue("id", "sess-a")
	req.SetPathValue("request_id", requestID)
	req.Header.Set("Authorization", eventsBearerFor(t, pool, "user-alice"))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404 (non-owner rejected by requireSessionOwner): %s", rec.Code, rec.Body.String())
	}
}
