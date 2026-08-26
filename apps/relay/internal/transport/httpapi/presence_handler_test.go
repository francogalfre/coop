package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
)

func TestHandlePresenceShapeAndWindowFiltering(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	now := time.Now()

	registry.SessionStarted("sess-a", "/repo", "Alice", "claude-code", now)
	if err := registry.FileTouched("sess-a", "src/foo.ts", "write", now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := registry.FileTouched("sess-a", "src/old.ts", "read", now.Add(-2*time.Hour)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := pool.CreateAgentSession(t.Context(), "sess-a", proj, "user-owner", "/repo", "/repo", "claude-code", now); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/presence?repo=/repo&paths=src/foo.ts,src/bar.ts,src/old.ts&window_seconds=900", nil)
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handlePresence(pool, registry)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var payload presenceResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if payload.Repo != "/repo" || payload.WindowSeconds != 900 {
		t.Fatalf("unexpected header fields: %+v", payload)
	}
	if len(payload.Paths) != 3 {
		t.Fatalf("expected 3 path entries, got %d", len(payload.Paths))
	}

	if payload.Paths[0].Path != "src/foo.ts" || len(payload.Paths[0].Signals) != 1 {
		t.Fatalf("unexpected src/foo.ts entry: %+v", payload.Paths[0])
	}
	if payload.Paths[0].Signals[0].SessionID != "sess-a" || payload.Paths[0].Signals[0].Mode != "write" {
		t.Fatalf("unexpected signal: %+v", payload.Paths[0].Signals[0])
	}

	if payload.Paths[1].Path != "src/bar.ts" || len(payload.Paths[1].Signals) != 0 {
		t.Fatalf("expected empty signals for src/bar.ts, got %+v", payload.Paths[1])
	}

	if payload.Paths[2].Path != "src/old.ts" || len(payload.Paths[2].Signals) != 0 {
		t.Fatalf("expected src/old.ts filtered out by window, got %+v", payload.Paths[2])
	}
}

func TestHandlePresenceFiltersOutNonMemberSessions(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	otherProj, err := pool.CreateProject(t.Context(), "Other", "other", "user-other")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	now := time.Now()

	registry.SessionStarted("sess-b", "/repo", "Bob", "opencode", now)
	if err := registry.FileTouched("sess-b", "src/foo.ts", "write", now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := pool.CreateAgentSession(t.Context(), "sess-b", otherProj, "user-other", "/repo", "/repo", "opencode", now); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/presence?repo=/repo&paths=src/foo.ts", nil)
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handlePresence(pool, registry)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var payload presenceResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(payload.Paths) != 1 || len(payload.Paths[0].Signals) != 0 {
		t.Fatalf("expected non-member session's signal filtered out, got %+v", payload.Paths)
	}
}

func TestHandlePresenceMissingRepo(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	registry := presence.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/presence?paths=src/foo.ts", nil)
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handlePresence(pool, registry)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestHandlePresenceMissingPaths(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	registry := presence.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/presence?repo=/repo", nil)
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handlePresence(pool, registry)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestHandlePresenceRequiresIdentity(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	registry := presence.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/presence?repo=/repo&paths=src/foo.ts", nil)
	rec := httptest.NewRecorder()

	handlePresence(pool, registry)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}
