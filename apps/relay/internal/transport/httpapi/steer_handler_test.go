package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func withActorNamed(req *http.Request, userID, displayName string) *http.Request {
	return req.WithContext(auth.WithActor(req.Context(), auth.Actor{UserID: userID, DisplayName: displayName}))
}

func steerSessionFixture(t *testing.T, sessionID string) *db.Pool {
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

func doSteerPost(t *testing.T, pool *db.Pool, mailbox *stream.Mailbox, store *stream.Store, sessionID, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/steer", strings.NewReader(body))
	req.SetPathValue("id", sessionID)
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handleSteerPost(pool, mailbox, store, stream.NewSteerRequestRegistry())(rec, req)

	return rec
}

func doSteerGet(t *testing.T, mailbox *stream.Mailbox, sessionID string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/"+sessionID+"/steer", nil)
	req.SetPathValue("id", sessionID)
	rec := httptest.NewRecorder()

	handleSteerGet(mailbox, stream.NewTakeoverRegistry())(rec, req)

	return rec
}

func TestSteerPostThenGetOnceThenEmpty(t *testing.T) {
	pool := steerSessionFixture(t, "sess-a")
	mailbox := stream.NewMailbox()
	store := stream.New()

	postRec := doSteerPost(t, pool, mailbox, store, "sess-a", `{"text":"try the other branch"}`)
	if postRec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", postRec.Code, postRec.Body.String())
	}

	getRec := doSteerGet(t, mailbox, "sess-a")
	if getRec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", getRec.Code, getRec.Body.String())
	}

	var payload steerGetResponse
	if err := json.NewDecoder(getRec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if !payload.HasMessage || payload.From != "Alice" || payload.Text != "try the other branch" {
		t.Fatalf("unexpected payload: %+v", payload)
	}

	secondGetRec := doSteerGet(t, mailbox, "sess-a")
	if secondGetRec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 on second get: %s", secondGetRec.Code, secondGetRec.Body.String())
	}

	var secondPayload steerGetResponse
	if err := json.NewDecoder(secondGetRec.Body).Decode(&secondPayload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if secondPayload.HasMessage {
		t.Fatalf("got has_message true on second get, want false (mailbox drained): %+v", secondPayload)
	}
}

func TestSteerGetUnknownSessionHasNoMessage(t *testing.T) {
	mailbox := stream.NewMailbox()

	rec := doSteerGet(t, mailbox, "sess-ghost")
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}

	var payload steerGetResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.HasMessage {
		t.Fatalf("got has_message true for unknown session, want false: %+v", payload)
	}
}

func TestHandleSteerPostRequiresActor(t *testing.T) {
	mailbox := stream.NewMailbox()
	store := stream.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/steer", strings.NewReader(`{"text":"hello"}`))
	req.SetPathValue("id", "sess-a")
	rec := httptest.NewRecorder()

	handleSteerPost(nil, mailbox, store, stream.NewSteerRequestRegistry())(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleSteerPostValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing text", `{}`},
		{"empty text", `{"text":""}`},
		{"malformed JSON", `{not json`},
	}

	mailbox := stream.NewMailbox()
	store := stream.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doSteerPost(t, nil, mailbox, store, "sess-a", tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestSteerPostEchoesHumanSteerToStore(t *testing.T) {
	pool := steerSessionFixture(t, "sess-a")
	mailbox := stream.NewMailbox()
	store := stream.New()

	rec := doSteerPost(t, pool, mailbox, store, "sess-a", `{"text":"try the other branch"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	var payload steerPostResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.Queued != 1 {
		t.Fatalf("got queued %d, want 1", payload.Queued)
	}

	events := store.Since("sess-a", 0)
	if len(events) != 1 {
		t.Fatalf("got %d events in store, want 1", len(events))
	}

	var fields map[string]any
	if err := json.Unmarshal(events[0].Data, &fields); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}

	if fields["type"] != "human.steer" {
		t.Fatalf("got type %v, want human.steer", fields["type"])
	}

	actor, _ := fields["actor"].(map[string]any)
	if actor["display_name"] != "Alice" {
		t.Fatalf("got actor %+v, want display_name Alice", actor)
	}

	text, _ := fields["text"].(map[string]any)
	if text["text"] != "try the other branch" {
		t.Fatalf("got text %+v, want \"try the other branch\"", text)
	}

	if fields["seq"] != float64(1) {
		t.Fatalf("got seq %v, want the DB-assigned seq 1 (not the hardcoded 0)", fields["seq"])
	}
}

func TestSteerPostPersistsToPostgresWithMatchingSeq(t *testing.T) {
	pool := steerSessionFixture(t, "sess-a")
	mailbox := stream.NewMailbox()
	store := stream.New()

	rec := doSteerPost(t, pool, mailbox, store, "sess-a", `{"text":"try the other branch"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	dbEvents, err := pool.EventsSince(t.Context(), "sess-a", 0, 10)
	if err != nil {
		t.Fatalf("EventsSince: %v", err)
	}
	if len(dbEvents) != 1 {
		t.Fatalf("expected the steer message to be persisted, got %d events", len(dbEvents))
	}

	memEvents := store.Since("sess-a", 0)
	if len(memEvents) != 1 {
		t.Fatalf("got %d events in store, want 1", len(memEvents))
	}

	if dbEvents[0].Seq != memEvents[0].Seq {
		t.Fatalf("postgres seq %d disagrees with in-memory store seq %d", dbEvents[0].Seq, memEvents[0].Seq)
	}
}

func TestSteerPostIgnoresForgedFromInBody(t *testing.T) {
	pool := steerSessionFixture(t, "sess-a")
	mailbox := stream.NewMailbox()
	store := stream.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/steer", strings.NewReader(`{"from":"Not Alice","text":"try the other branch"}`))
	req.SetPathValue("id", "sess-a")
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handleSteerPost(pool, mailbox, store, stream.NewSteerRequestRegistry())(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	getRec := doSteerGet(t, mailbox, "sess-a")

	var payload steerGetResponse
	if err := json.NewDecoder(getRec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.From != "Alice" {
		t.Fatalf("got from %q, want %q (authenticated actor, not the forged body value)", payload.From, "Alice")
	}

	events := store.Since("sess-a", 0)
	var fields map[string]any
	if err := json.Unmarshal(events[0].Data, &fields); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	actor, _ := fields["actor"].(map[string]any)
	if actor["display_name"] != "Alice" {
		t.Fatalf("got actor %+v, want display_name Alice", actor)
	}
}
