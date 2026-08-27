package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func TestHandleIngestRejectsEventWithoutIdentity(t *testing.T) {
	registry := presence.New()
	store := stream.New()

	body := `{"v":1,"session_id":"sess-a","seq":0,"ts":"2026-08-24T10:00:00Z","type":"agent.text"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handleIngest(nil, registry, store)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}

	if events := store.Since("sess-a", 0); len(events) != 0 {
		t.Fatalf("expected unauthenticated event not appended, got %d", len(events))
	}
}

func TestHandleIngestRejectsNonOwnerForExistingSession(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	store := stream.New()
	ingest := handleIngest(pool, registry, store)

	now := time.Now().UTC()
	startBody := `{"v":1,"session_id":"sess-forge","seq":0,"ts":"` + now.Format(time.RFC3339) + `","type":"session.start","cwd":"/repo","owner":{"id":"alice","display_name":"Alice"}}`

	startReq := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(startBody))
	startReq.Header.Set(coopProjectHeader, "coop")
	startReq = withActor(startReq, "user-owner")
	startRec := httptest.NewRecorder()

	ingest(startRec, startReq)
	if startRec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", startRec.Code, startRec.Body.String())
	}

	forgedBody := `{"v":1,"session_id":"sess-forge","seq":1,"ts":"` + now.Add(time.Second).Format(time.RFC3339) + `","type":"agent.text"}`

	forgedReq := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(forgedBody))
	forgedReq.Header.Set(coopProjectHeader, "coop")
	forgedReq = withActor(forgedReq, "user-intruder")
	forgedRec := httptest.NewRecorder()

	ingest(forgedRec, forgedReq)

	if forgedRec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", forgedRec.Code, forgedRec.Body.String())
	}

	events, err := pool.EventsSince(t.Context(), "sess-forge", 0, 10)
	if err != nil {
		t.Fatalf("EventsSince: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected only the session.start event persisted, got %d", len(events))
	}
}

func TestHandleIngestOwnerCanSendEveryEventType(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	store := stream.New()
	ingest := handleIngest(pool, registry, store)

	now := time.Now().UTC()

	bodies := []string{
		`{"v":1,"session_id":"sess-owned","seq":0,"ts":"` + now.Format(time.RFC3339) + `","type":"session.start","cwd":"/repo","owner":{"id":"alice","display_name":"Alice"}}`,
		`{"v":1,"session_id":"sess-owned","seq":1,"ts":"` + now.Add(time.Second).Format(time.RFC3339) + `","type":"agent.text"}`,
		`{"v":1,"session_id":"sess-owned","seq":2,"ts":"` + now.Add(2*time.Second).Format(time.RFC3339) + `","type":"file.touched","path":"src/foo.ts","mode":"write"}`,
		`{"v":1,"session_id":"sess-owned","seq":3,"ts":"` + now.Add(3*time.Second).Format(time.RFC3339) + `","type":"session.end"}`,
	}

	for _, body := range bodies {
		req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(body))
		req.Header.Set(coopProjectHeader, "coop")
		req = withActor(req, "user-owner")
		rec := httptest.NewRecorder()

		ingest(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
		}
	}
}
