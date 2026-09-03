package httpapi

import (
	"encoding/hex"
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

func bearerForMessage(t *testing.T, pool *db.Pool, userID string) string {
	t.Helper()

	rawToken, err := pool.CreateCliCredential(t.Context(), userID, "Display "+userID)
	if err != nil {
		t.Fatalf("CreateCliCredential: %v", err)
	}

	return "Bearer " + hex.EncodeToString(rawToken)
}

func messageSessionFixture(t *testing.T, sessionID string) *db.Pool {
	t.Helper()

	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-alice")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := pool.CreateAgentSession(t.Context(), sessionID, proj, "user-alice", "/repo", "/repo", "claude-code", time.Now()); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	return pool
}

func doMessagePost(t *testing.T, pool *db.Pool, store *stream.Store, sessionID, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/message", strings.NewReader(body))
	req.SetPathValue("id", sessionID)
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handleMessagePost(pool, store)(rec, req)

	return rec
}

func TestHandleMessagePostRequiresActor(t *testing.T) {
	store := stream.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/message", strings.NewReader(`{"text":"hello"}`))
	req.SetPathValue("id", "sess-a")
	rec := httptest.NewRecorder()

	handleMessagePost(nil, store)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleMessagePostNonMemberRejected(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	store := stream.New()
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/message", strings.NewReader(`{"text":"hello"}`))
	req.SetPathValue("id", "sess-a")
	req.Header.Set("Authorization", bearerForMessage(t, pool, "user-stranger"))
	rec := httptest.NewRecorder()

	requireSessionMember := auth.RequireSessionMember(pool, cfg)
	requireSessionMember(handleMessagePost(pool, store))(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleMessagePostValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing text", `{}`},
		{"empty text", `{"text":""}`},
		{"malformed JSON", `{not json`},
		{"oversized text", `{"text":"` + strings.Repeat("a", messageTextMax+1) + `"}`},
		{"negative anchor_seq", `{"text":"hi","anchor_seq":-1}`},
	}

	store := stream.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doMessagePost(t, nil, store, "sess-a", tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestMessagePostPersistsAndAppearsInEvents(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	store := stream.New()

	rec := doMessagePost(t, pool, store, "sess-a", `{"text":"worth a look before this lands"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	var payload messagePostResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.Status != "sent" {
		t.Fatalf("got status %q, want %q", payload.Status, "sent")
	}
	if payload.Seq != 1 {
		t.Fatalf("got seq %d, want 1 (the DB-assigned seq)", payload.Seq)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/sessions/sess-a/events", nil)
	getReq.SetPathValue("id", "sess-a")
	getRec := httptest.NewRecorder()
	handleEvents(pool)(getRec, getReq)

	var eventsPayload eventsResponse
	if err := json.NewDecoder(getRec.Body).Decode(&eventsPayload); err != nil {
		t.Fatalf("failed to decode events body: %v", err)
	}
	if len(eventsPayload.Events) != 1 {
		t.Fatalf("got %d events, want 1", len(eventsPayload.Events))
	}

	var fields map[string]any
	if err := json.Unmarshal(eventsPayload.Events[0], &fields); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if fields["type"] != "human.message" {
		t.Fatalf("got type %v, want human.message", fields["type"])
	}
	actor, _ := fields["actor"].(map[string]any)
	if actor["display_name"] != "Alice" {
		t.Fatalf("got actor %+v, want display_name Alice", actor)
	}
	if _, hasAnchor := fields["anchor_seq"]; hasAnchor {
		t.Fatalf("got anchor_seq present on an unanchored message: %+v", fields)
	}
}

func TestMessagePostRoundTripsAnchorSeq(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	store := stream.New()

	rec := doMessagePost(t, pool, store, "sess-a", `{"text":"why this tool call?","anchor_seq":12}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	events := store.Since("sess-a", 0)
	if len(events) != 1 {
		t.Fatalf("got %d events in store, want 1", len(events))
	}

	var fields map[string]any
	if err := json.Unmarshal(events[0].Data, &fields); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if fields["anchor_seq"] != float64(12) {
		t.Fatalf("got anchor_seq %v, want 12", fields["anchor_seq"])
	}
}

func TestMessagePostRoundTripsClientID(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	store := stream.New()

	rec := doMessagePost(t, pool, store, "sess-a", `{"text":"hi","client_id":"c-1"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	events := store.Since("sess-a", 0)
	var fields map[string]any
	if err := json.Unmarshal(events[0].Data, &fields); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if fields["client_id"] != "c-1" {
		t.Fatalf("got client_id %v, want c-1", fields["client_id"])
	}
}
