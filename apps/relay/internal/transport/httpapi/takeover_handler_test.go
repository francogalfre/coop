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

func takeoverSessionFixture(t *testing.T, sessionID string) *db.Pool {
	t.Helper()

	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if err := pool.AddMember(t.Context(), proj, "user-alice", db.RoleMember); err != nil {
		t.Fatalf("AddMember alice: %v", err)
	}
	if err := pool.AddMember(t.Context(), proj, "user-bob", db.RoleMember); err != nil {
		t.Fatalf("AddMember bob: %v", err)
	}

	if _, err := pool.CreateAgentSession(t.Context(), sessionID, proj, "user-owner", "/repo", "/repo", "claude-code", time.Now()); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	return pool
}

func doTakeoverPost(t *testing.T, pool *db.Pool, store *stream.Store, registry *stream.TakeoverRegistry, sessionID, userID, displayName, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/takeover", strings.NewReader(body))
	req.SetPathValue("id", sessionID)
	req = withActorNamed(req, userID, displayName)
	rec := httptest.NewRecorder()

	handleTakeoverPost(pool, store, registry)(rec, req)

	return rec
}

func TestTakeoverPostClaimThenRelease(t *testing.T) {
	pool := takeoverSessionFixture(t, "sess-a")
	store := stream.New()
	registry := stream.NewTakeoverRegistry(pool)

	claimRec := doTakeoverPost(t, pool, store, registry, "sess-a", "user-alice", "Alice", `{"active":true}`)
	if claimRec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", claimRec.Code, claimRec.Body.String())
	}

	state, err := registry.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !state.Active || state.By != "Alice" {
		t.Fatalf("got registry state %+v, want active held by Alice", state)
	}

	releaseRec := doTakeoverPost(t, pool, store, registry, "sess-a", "user-alice", "Alice", `{"active":false}`)
	if releaseRec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", releaseRec.Code, releaseRec.Body.String())
	}

	got, err := registry.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Active {
		t.Fatalf("got registry state %+v, want inactive after release", got)
	}
}

func TestTakeoverPostRejectsSecondActorWhileHeld(t *testing.T) {
	pool := takeoverSessionFixture(t, "sess-a")
	store := stream.New()
	registry := stream.NewTakeoverRegistry(pool)

	doTakeoverPost(t, pool, store, registry, "sess-a", "user-alice", "Alice", `{"active":true}`)

	rec := doTakeoverPost(t, pool, store, registry, "sess-a", "user-bob", "Bob", `{"active":true}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("got status %d, want 409: %s", rec.Code, rec.Body.String())
	}

	var payload takeoverResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.By != "Alice" {
		t.Fatalf("got by %q, want Alice", payload.By)
	}

	got, err := registry.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Active || got.By != "Alice" {
		t.Fatalf("got registry state %+v, want still held by Alice", got)
	}
}

func TestTakeoverPostReleaseRejectedForNonHolderNonOwner(t *testing.T) {
	pool := takeoverSessionFixture(t, "sess-a")
	store := stream.New()
	registry := stream.NewTakeoverRegistry(pool)

	doTakeoverPost(t, pool, store, registry, "sess-a", "user-alice", "Alice", `{"active":true}`)

	rec := doTakeoverPost(t, pool, store, registry, "sess-a", "user-bob", "Bob", `{"active":false}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("got status %d, want 403: %s", rec.Code, rec.Body.String())
	}

	got, err := registry.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Active || got.By != "Alice" {
		t.Fatalf("got registry state %+v, want still held by Alice", got)
	}
}

func TestTakeoverPostSessionOwnerCanForceRelease(t *testing.T) {
	pool := takeoverSessionFixture(t, "sess-a")
	store := stream.New()
	registry := stream.NewTakeoverRegistry(pool)

	doTakeoverPost(t, pool, store, registry, "sess-a", "user-alice", "Alice", `{"active":true}`)

	rec := doTakeoverPost(t, pool, store, registry, "sess-a", "user-owner", "Owner", `{"active":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	got, err := registry.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Active {
		t.Fatalf("got registry state %+v, want inactive after owner force-release", got)
	}
}

func TestTakeoverPostPersistsEventToPostgresAndStore(t *testing.T) {
	pool := takeoverSessionFixture(t, "sess-a")
	store := stream.New()
	registry := stream.NewTakeoverRegistry(pool)

	rec := doTakeoverPost(t, pool, store, registry, "sess-a", "user-alice", "Alice", `{"active":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	dbEvents, err := pool.EventsSince(t.Context(), "sess-a", 0, 10)
	if err != nil {
		t.Fatalf("EventsSince: %v", err)
	}
	if len(dbEvents) != 1 {
		t.Fatalf("got %d persisted events, want 1", len(dbEvents))
	}

	var fields map[string]any
	if err := json.Unmarshal(dbEvents[0].Data, &fields); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if fields["type"] != "human.takeover" || fields["active"] != true {
		t.Fatalf("got fields %+v, want type human.takeover and active true", fields)
	}

	memEvents := store.Since("sess-a", 0)
	if len(memEvents) != 1 || memEvents[0].Seq != dbEvents[0].Seq {
		t.Fatalf("got in-memory events %+v, want one matching the DB-assigned seq %d", memEvents, dbEvents[0].Seq)
	}
}

func TestTakeoverPostRequiresActor(t *testing.T) {
	store := stream.New()
	registry := stream.NewTakeoverRegistry(nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/takeover", strings.NewReader(`{"active":true}`))
	req.SetPathValue("id", "sess-a")
	rec := httptest.NewRecorder()

	handleTakeoverPost(nil, store, registry)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestSteerGetReflectsTakeoverState(t *testing.T) {
	mailbox := stream.NewMailbox()
	registry := stream.NewTakeoverRegistry(nil)
	if err := registry.Set(t.Context(), "sess-a", stream.TakeoverState{Active: true, ByID: "u1", By: "Alice"}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/sess-a/steer", nil)
	req.SetPathValue("id", "sess-a")
	rec := httptest.NewRecorder()

	handleSteerGet(mailbox, registry)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var payload steerGetResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.HasMessage {
		t.Fatal("got has_message true, want false (no mailbox message queued)")
	}
	if !payload.Takeover.Active || payload.Takeover.By != "Alice" {
		t.Fatalf("got takeover %+v, want active held by Alice", payload.Takeover)
	}
}
