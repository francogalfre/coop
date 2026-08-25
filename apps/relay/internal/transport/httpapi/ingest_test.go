package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/presence"
)

func doIngest(t *testing.T, registry *presence.Registry, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handleIngest(registry)(rec, req)

	return rec
}

func TestHandleIngestValidEvents(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "session.start",
			body: `{"v":1,"session_id":"sess-a","seq":0,"ts":"2026-08-24T10:00:00Z","type":"session.start","cwd":"/repo","owner":{"id":"alice","display_name":"Alice"}}`,
		},
		{
			name: "file.touched",
			body: `{"v":1,"session_id":"sess-a","seq":1,"ts":"2026-08-24T10:00:05Z","type":"file.touched","path":"src/foo.ts","mode":"write"}`,
		},
		{
			name: "session.end",
			body: `{"v":1,"session_id":"sess-a","seq":2,"ts":"2026-08-24T10:05:00Z","type":"session.end"}`,
		},
	}

	registry := presence.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doIngest(t, registry, tt.body)
			if rec.Code != http.StatusAccepted {
				t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
			}
		})
	}

	signals := registry.Query("/repo", []string{"src/foo.ts"}, time.Now(), 24*time.Hour)
	if len(signals["src/foo.ts"]) != 1 {
		t.Fatalf("expected registry mutated with 1 signal, got %d", len(signals["src/foo.ts"]))
	}

	active := registry.ActiveSessions("/repo")
	if len(active) != 0 {
		t.Fatalf("expected session ended, got %d active", len(active))
	}
}

func TestHandleIngestFractionalSecondsTimestamp(t *testing.T) {
	registry := presence.New()

	body := `{"v":1,"session_id":"sess-a","seq":0,"ts":"2026-08-24T15:31:07.812Z","type":"session.start","cwd":"/repo","owner":{"id":"alice","display_name":"Alice"}}`

	rec := doIngest(t, registry, body)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleIngestMalformedJSON(t *testing.T) {
	registry := presence.New()

	rec := doIngest(t, registry, `{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestHandleIngestMissingRequiredEnvelopeField(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing v", `{"session_id":"sess-a","seq":0,"ts":"2026-08-24T10:00:00Z","type":"session.end"}`},
		{"missing session_id", `{"v":1,"seq":0,"ts":"2026-08-24T10:00:00Z","type":"session.end"}`},
		{"negative seq", `{"v":1,"session_id":"sess-a","seq":-1,"ts":"2026-08-24T10:00:00Z","type":"session.end"}`},
		{"bad ts", `{"v":1,"session_id":"sess-a","seq":0,"ts":"not-a-timestamp","type":"session.end"}`},
	}

	registry := presence.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doIngest(t, registry, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleIngestFileTouchedUnknownSession(t *testing.T) {
	registry := presence.New()

	body := `{"v":1,"session_id":"sess-ghost","seq":0,"ts":"2026-08-24T10:00:00Z","type":"file.touched","path":"src/foo.ts","mode":"write"}`

	rec := doIngest(t, registry, body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
	}

	var payload map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if payload["error"] == "" {
		t.Fatal("expected error message in body")
	}
}

func TestHandleIngestUnknownEventTypeNoOp(t *testing.T) {
	registry := presence.New()

	body := `{"v":1,"session_id":"sess-a","seq":0,"ts":"2026-08-24T10:00:00Z","type":"tool.blocked"}`

	rec := doIngest(t, registry, body)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	active := registry.ActiveSessions("/repo")
	if len(active) != 0 {
		t.Fatalf("expected registry unchanged, got %d active sessions", len(active))
	}
}
