package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/presence"
)

func TestHandleSessionsShape(t *testing.T) {
	registry := presence.New()
	now := time.Now()

	registry.SessionStarted("sess-a", "/repo", "Alice", now)
	registry.SessionStarted("sess-b", "/repo", "Bob", now)
	registry.SessionEnded("sess-b", now)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions?repo=/repo", nil)
	rec := httptest.NewRecorder()

	handleSessions(registry)(rec, req)

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
		t.Fatalf("expected 1 active session, got %d", len(payload.Sessions))
	}
	if payload.Sessions[0].SessionID != "sess-a" || payload.Sessions[0].Owner != "Alice" {
		t.Fatalf("unexpected session: %+v", payload.Sessions[0])
	}
	if !payload.Sessions[0].Active {
		t.Fatal("expected session to be active")
	}
}

func TestHandleSessionsMissingRepo(t *testing.T) {
	registry := presence.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions", nil)
	rec := httptest.NewRecorder()

	handleSessions(registry)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}
