package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func TestHandleViewersReturnsCurrentViewers(t *testing.T) {
	hub := stream.NewPresenceHub()
	hub.AddViewer("sess-a", "Alice")
	hub.AddViewer("sess-a", "Bob")

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/sess-a/viewers", nil)
	req.SetPathValue("id", "sess-a")
	rec := httptest.NewRecorder()

	handleViewers(hub)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var payload viewersResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(payload.Viewers) != 2 {
		t.Fatalf("got %v, want 2 viewers", payload.Viewers)
	}
}

func TestHandleViewersReturnsEmptyForUnknownSession(t *testing.T) {
	hub := stream.NewPresenceHub()

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/sess-ghost/viewers", nil)
	req.SetPathValue("id", "sess-ghost")
	rec := httptest.NewRecorder()

	handleViewers(hub)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var payload viewersResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(payload.Viewers) != 0 {
		t.Fatalf("got %v, want no viewers", payload.Viewers)
	}
}
