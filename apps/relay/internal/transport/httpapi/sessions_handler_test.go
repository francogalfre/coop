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

func TestHandleSessionsShapeFiltersToMemberProjects(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	otherProj, err := pool.CreateProject(t.Context(), "Other", "other", "user-other")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	now := time.Now()

	registry.SessionStarted("sess-a", "/repo", "Alice", "claude-code", now)
	registry.SessionStarted("sess-b", "/repo", "Bob", "opencode", now)

	if _, err := pool.CreateAgentSession(t.Context(), "sess-a", proj, "user-owner", "/repo", "/repo", "claude-code", now); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}
	if _, err := pool.CreateAgentSession(t.Context(), "sess-b", otherProj, "user-other", "/repo", "/repo", "opencode", now); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions?repo=/repo", nil)
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handleSessions(pool, registry)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var payload sessionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if payload.Repo != "/repo" {
		t.Fatalf("unexpected repo: %s", payload.Repo)
	}
	if len(payload.Sessions) != 1 {
		t.Fatalf("expected 1 active session visible to user-owner, got %d: %+v", len(payload.Sessions), payload.Sessions)
	}
	if payload.Sessions[0].SessionID != "sess-a" {
		t.Fatalf("unexpected session: %+v", payload.Sessions[0])
	}
}

func TestHandleSessionsMissingRepo(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	registry := presence.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions", nil)
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handleSessions(pool, registry)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestHandleSessionsRequiresIdentity(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	registry := presence.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions?repo=/repo", nil)
	rec := httptest.NewRecorder()

	handleSessions(pool, registry)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}
