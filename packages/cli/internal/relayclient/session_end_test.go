package relayclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

func TestPostSessionEndSendsSessionEndEnvelope(t *testing.T) {
	var got map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	}))
	defer server.Close()

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-xyz", CLICredential: "tok", Project: "demo"}

	if err := PostSessionEnd(context.Background(), cfg); err != nil {
		t.Fatalf("PostSessionEnd: %v", err)
	}

	if got["type"] != "session.end" {
		t.Fatalf("type = %v, want session.end", got["type"])
	}
	if got["session_id"] != "sess-xyz" {
		t.Fatalf("session_id = %v", got["session_id"])
	}
	if got["v"] != float64(1) {
		t.Fatalf("v = %v, want 1", got["v"])
	}
}

func TestPostSessionEndReturnsErrorOnUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-xyz"}

	if err := PostSessionEnd(context.Background(), cfg); err == nil {
		t.Fatal("expected error on 404")
	}
}
