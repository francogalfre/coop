package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func TestHandlePermissionResolvePostEmitsResolution(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	store := stream.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/permissions/req-a/resolve", strings.NewReader(`{"decision":"deny","reason":"looks destructive"}`))
	req.SetPathValue("id", "sess-a")
	req.SetPathValue("request_id", "req-a")
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handlePermissionResolvePost(pool, store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	events := store.Since("sess-a", 0)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	var event struct {
		Type       string `json:"type"`
		RequestID  string `json:"request_id"`
		Decision   string `json:"decision"`
		ResolvedBy struct {
			DisplayName string `json:"display_name"`
		} `json:"resolved_by"`
	}
	if err := json.Unmarshal(events[0].Data, &event); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if event.Type != "permission.resolved" || event.RequestID != "req-a" || event.Decision != "deny" || event.ResolvedBy.DisplayName != "Alice" {
		t.Fatalf("got %+v, want a deny resolution attributed to Alice", event)
	}
}

func TestHandlePermissionResolvePostRejectsUnknownDecision(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	store := stream.New()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/permissions/req-a/resolve", strings.NewReader(`{"decision":"maybe"}`))
	req.SetPathValue("id", "sess-a")
	req.SetPathValue("request_id", "req-a")
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handlePermissionResolvePost(pool, store)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}
