package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func modeSessionFixture(t *testing.T, sessionID, ownerID string) *db.Pool {
	t.Helper()

	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", ownerID)
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := pool.CreateAgentSession(t.Context(), sessionID, proj, ownerID, "/repo", "/repo", "claude-code", time.Now()); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	return pool
}

func doModePost(t *testing.T, pool *db.Pool, store *stream.Store, sessionID, userID, displayName, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/mode", strings.NewReader(body))
	req.SetPathValue("id", sessionID)
	req = withActorNamed(req, userID, displayName)
	rec := httptest.NewRecorder()

	handleModePost(pool, store)(rec, req)

	return rec
}

func TestModePostByOwnerPersistsAndBroadcasts(t *testing.T) {
	pool := modeSessionFixture(t, "sess-a", "user-owner")
	store := stream.New()

	rec := doModePost(t, pool, store, "sess-a", "user-owner", "Owner", `{"mode":"restricted"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var payload sessionModePostResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.Mode != "restricted" {
		t.Fatalf("got mode %q, want restricted", payload.Mode)
	}

	sess, err := pool.GetAgentSession(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("GetAgentSession: %v", err)
	}
	if sess.Mode != "restricted" {
		t.Fatalf("got persisted mode %q, want restricted", sess.Mode)
	}

	events := store.Since("sess-a", 0)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	var fields map[string]any
	if err := json.Unmarshal(events[0].Data, &fields); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if fields["type"] != "session.mode_changed" || fields["mode"] != "restricted" {
		t.Fatalf("got event %+v, want session.mode_changed with mode restricted", fields)
	}
}

func TestModePostInvalidValueRejected(t *testing.T) {
	pool := modeSessionFixture(t, "sess-a", "user-owner")
	store := stream.New()

	rec := doModePost(t, pool, store, "sess-a", "user-owner", "Owner", `{"mode":"chaotic"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func TestModePostRouteRejectsNonOwner(t *testing.T) {
	pool := modeSessionFixture(t, "sess-a", "user-owner")
	store := stream.New()
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	handler := auth.RequireSessionOwner(pool, cfg)(handleModePost(pool, store))

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/mode", strings.NewReader(`{"mode":"restricted"}`))
	req.SetPathValue("id", "sess-a")
	req.Header.Set("Authorization", eventsBearerFor(t, pool, "user-alice"))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404 (non-owner rejected by requireSessionOwner): %s", rec.Code, rec.Body.String())
	}
}
