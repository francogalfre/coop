package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func TestHandleIngestPersistsSessionWithProjectAndMembership(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	store := stream.New()
	ingest := handleIngest(pool, registry, store)

	now := time.Now().UTC()

	startBody := `{"v":1,"session_id":"sess-persisted","seq":0,"ts":"` + now.Format(time.RFC3339) + `","type":"session.start","cwd":"/repo","owner":{"id":"alice","display_name":"Alice"},"harness":"claude-code"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(startBody))
	req.Header.Set(coopProjectHeader, "coop")
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	ingest(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	sess, err := pool.GetAgentSession(t.Context(), "sess-persisted")
	if err != nil {
		t.Fatalf("GetAgentSession: %v", err)
	}

	if sess.Status != db.SessionStatusLive {
		t.Fatalf("got status %q, want %q", sess.Status, db.SessionStatusLive)
	}
	if sess.Harness != "claude-code" {
		t.Fatalf("got harness %q, want %q", sess.Harness, "claude-code")
	}
	if sess.OwnerID != "user-owner" {
		t.Fatalf("got owner_id %q, want %q", sess.OwnerID, "user-owner")
	}

	touchedBody := `{"v":1,"session_id":"sess-persisted","seq":1,"ts":"` + now.Add(time.Second).Format(time.RFC3339) + `","type":"file.touched","path":"src/foo.ts","mode":"write"}`

	req2 := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(touchedBody))
	req2.Header.Set(coopProjectHeader, "coop")
	rec2 := httptest.NewRecorder()

	ingest(rec2, req2)

	if rec2.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec2.Code, rec2.Body.String())
	}

	events, err := pool.EventsSince(t.Context(), "sess-persisted", 0, 10)
	if err != nil {
		t.Fatalf("EventsSince: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 persisted event, got %d", len(events))
	}

	endBody := `{"v":1,"session_id":"sess-persisted","seq":2,"ts":"` + now.Add(2*time.Second).Format(time.RFC3339) + `","type":"session.end"}`

	req3 := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(endBody))
	req3.Header.Set(coopProjectHeader, "coop")
	rec3 := httptest.NewRecorder()

	ingest(rec3, req3)

	if rec3.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec3.Code, rec3.Body.String())
	}

	ended, err := pool.GetAgentSession(t.Context(), "sess-persisted")
	if err != nil {
		t.Fatalf("GetAgentSession: %v", err)
	}
	if ended.Status != db.SessionStatusEnded {
		t.Fatalf("got status %q, want %q", ended.Status, db.SessionStatusEnded)
	}
	if ended.EndedAt == nil {
		t.Fatal("expected ended_at to be set")
	}
}

func TestHandleIngestSessionStartRejectsNonMember(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	store := stream.New()

	body := `{"v":1,"session_id":"sess-x","seq":0,"ts":"2026-08-24T10:00:00Z","type":"session.start","cwd":"/repo","owner":{"id":"alice","display_name":"Alice"}}`

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(body))
	req.Header.Set(coopProjectHeader, "coop")
	req = withActor(req, "user-stranger")
	rec := httptest.NewRecorder()

	handleIngest(pool, registry, store)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}

	if _, err := pool.GetAgentSession(t.Context(), "sess-x"); err == nil {
		t.Fatal("expected no agent session to be created for a non-member")
	}
}

func TestHandleIngestAnonymousFlowTouchesNoPostgres(t *testing.T) {
	registry := presence.New()
	store := stream.New()

	now := time.Now().UTC()

	body := `{"v":1,"session_id":"sess-anon","seq":0,"ts":"` + now.Format(time.RFC3339) + `","type":"session.start","cwd":"/repo","owner":{"id":"alice","display_name":"Alice"}}`

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handleIngest(nil, registry, store)(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	events := store.Since("sess-anon", 0)
	if len(events) != 1 {
		t.Fatalf("expected 1 event appended to in-memory stream, got %d", len(events))
	}
}
