package httpapi

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func eventsBearerFor(t *testing.T, pool *db.Pool, userID string) string {
	t.Helper()

	rawToken, err := pool.CreateCliCredential(t.Context(), userID, "Display "+userID)
	if err != nil {
		t.Fatalf("CreateCliCredential: %v", err)
	}

	return "Bearer " + hex.EncodeToString(rawToken)
}

func eventsFixture(t *testing.T, n int) (pool *db.Pool, sessionID string) {
	t.Helper()

	pool = dbtest.OpenScratch(t)
	sessionID = "sess-events"

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if err := pool.AddMember(t.Context(), proj, "user-member", db.RoleMember); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	if _, err := pool.CreateAgentSession(t.Context(), sessionID, proj, "user-owner", "/repo", "/repo", "claude-code", time.Now()); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	for i := 0; i < n; i++ {
		if _, err := pool.AppendEvent(t.Context(), sessionID, []byte(`{"n":`+strconv.Itoa(i)+`}`)); err != nil {
			t.Fatalf("AppendEvent: %v", err)
		}
	}

	return pool, sessionID
}

func TestHandleEventsMemberGetsValidPage(t *testing.T) {
	pool, sessionID := eventsFixture(t, 5)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/"+sessionID+"/events", nil)
	req.SetPathValue("id", sessionID)
	rec := httptest.NewRecorder()

	handleEvents(pool)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var payload eventsResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(payload.Events) != 5 {
		t.Fatalf("got %d events, want 5", len(payload.Events))
	}
	if payload.OldestSeq != 1 {
		t.Fatalf("got oldest_seq %d, want 1", payload.OldestSeq)
	}
	if payload.HasMore {
		t.Fatal("expected has_more false when fewer events remain than the limit")
	}

	var first map[string]any
	if err := json.Unmarshal(payload.Events[0], &first); err != nil {
		t.Fatalf("failed to decode first event: %v", err)
	}
	if first["seq"] != float64(1) {
		t.Fatalf("got seq %v, want 1 (stamped to match the DB column)", first["seq"])
	}
}

func TestHandleEventsHasMoreTrueAtBoundary(t *testing.T) {
	pool, sessionID := eventsFixture(t, 5)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/"+sessionID+"/events?limit=3", nil)
	req.SetPathValue("id", sessionID)
	rec := httptest.NewRecorder()

	handleEvents(pool)(rec, req)

	var payload eventsResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(payload.Events) != 3 {
		t.Fatalf("got %d events, want 3", len(payload.Events))
	}
	if !payload.HasMore {
		t.Fatal("expected has_more true when exactly limit events are returned and more remain")
	}
	if payload.OldestSeq != 3 {
		t.Fatalf("got oldest_seq %d, want 3", payload.OldestSeq)
	}
}

func TestHandleEventsHasMoreFalseWhenExhausted(t *testing.T) {
	pool, sessionID := eventsFixture(t, 2)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/"+sessionID+"/events?limit=50", nil)
	req.SetPathValue("id", sessionID)
	rec := httptest.NewRecorder()

	handleEvents(pool)(rec, req)

	var payload eventsResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(payload.Events) != 2 {
		t.Fatalf("got %d events, want 2", len(payload.Events))
	}
	if payload.HasMore {
		t.Fatal("expected has_more false: only 2 events exist for a limit of 50")
	}
}

func TestHandleEventsBeforeCursorPagesBackward(t *testing.T) {
	pool, sessionID := eventsFixture(t, 5)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/"+sessionID+"/events?before=4&limit=2", nil)
	req.SetPathValue("id", sessionID)
	rec := httptest.NewRecorder()

	handleEvents(pool)(rec, req)

	var payload eventsResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(payload.Events) != 2 {
		t.Fatalf("got %d events, want 2", len(payload.Events))
	}

	var first, second map[string]any
	json.Unmarshal(payload.Events[0], &first)
	json.Unmarshal(payload.Events[1], &second)

	if first["seq"] != float64(2) || second["seq"] != float64(3) {
		t.Fatalf("got seqs %v, %v, want 2, 3 (the 2 events immediately before seq 4)", first["seq"], second["seq"])
	}
}

func TestHandleEventsRejectsInvalidBeforeAndLimit(t *testing.T) {
	pool, sessionID := eventsFixture(t, 1)

	tests := []string{
		"/v1/sessions/" + sessionID + "/events?before=-1",
		"/v1/sessions/" + sessionID + "/events?before=notanumber",
		"/v1/sessions/" + sessionID + "/events?limit=0",
		"/v1/sessions/" + sessionID + "/events?limit=-5",
		"/v1/sessions/" + sessionID + "/events?limit=notanumber",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.SetPathValue("id", sessionID)
			rec := httptest.NewRecorder()

			handleEvents(pool)(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleEventsLimitClampedAtMax(t *testing.T) {
	pool, sessionID := eventsFixture(t, 3)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/"+sessionID+"/events?limit=10000", nil)
	req.SetPathValue("id", sessionID)
	rec := httptest.NewRecorder()

	handleEvents(pool)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleEventsRouteMembershipGating(t *testing.T) {
	pool, sessionID := eventsFixture(t, 2)
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	handler := auth.RequireSessionMember(pool, cfg)(handleEvents(pool))

	for _, tt := range []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"member", eventsBearerFor(t, pool, "user-member"), http.StatusOK},
		{"owner", eventsBearerFor(t, pool, "user-owner"), http.StatusOK},
		{"non-member", eventsBearerFor(t, pool, "user-stranger"), http.StatusNotFound},
		{"unauthenticated", "", http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/sessions/"+sessionID+"/events", nil)
			req.SetPathValue("id", sessionID)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("%s: got status %d, want %d: %s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
