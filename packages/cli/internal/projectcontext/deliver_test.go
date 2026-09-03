package projectcontext

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

type fakeRelay struct {
	contextText    string
	contextVersion int
	contextStatus  int
	steered        []map[string]any
}

func (f *fakeRelay) server() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/projects/{slug}/context", func(w http.ResponseWriter, r *http.Request) {
		status := f.contextStatus
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"text":    f.contextText,
			"version": f.contextVersion,
		})
	})

	mux.HandleFunc("POST /v1/sessions/{id}/steer", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var decoded map[string]any
		_ = json.Unmarshal(body, &decoded)
		f.steered = append(f.steered, decoded)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	})

	return httptest.NewServer(mux)
}

func TestDeliverPostsFetchedContextAsSteer(t *testing.T) {
	relay := &fakeRelay{contextText: "read the runbook first", contextVersion: 3}
	server := relay.server()
	defer server.Close()

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-a", Project: "proj-a"}

	Deliver(context.Background(), cfg)

	if len(relay.steered) != 1 {
		t.Fatalf("got %d steer posts, want 1", len(relay.steered))
	}
	if relay.steered[0]["text"] != "read the runbook first" {
		t.Errorf("got text %v, want %q", relay.steered[0]["text"], "read the runbook first")
	}
	if relay.steered[0]["project_context_version"] != float64(3) {
		t.Errorf("got project_context_version %v, want 3", relay.steered[0]["project_context_version"])
	}
}

func TestDeliverSkipsWhenContextEmpty(t *testing.T) {
	relay := &fakeRelay{contextText: "", contextVersion: 0}
	server := relay.server()
	defer server.Close()

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-a", Project: "proj-a"}

	Deliver(context.Background(), cfg)

	if len(relay.steered) != 0 {
		t.Fatalf("got %d steer posts, want 0 for an unset project context", len(relay.steered))
	}
}

func TestDeliverSwallowsRelayError(t *testing.T) {
	relay := &fakeRelay{contextStatus: http.StatusInternalServerError}
	server := relay.server()
	defer server.Close()

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-a", Project: "proj-a"}

	Deliver(context.Background(), cfg)

	if len(relay.steered) != 0 {
		t.Fatalf("got %d steer posts, want 0 when the context fetch fails", len(relay.steered))
	}
}

func TestDeliverSkipsFetchWhenProjectEmpty(t *testing.T) {
	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-a", Project: ""}

	Deliver(context.Background(), cfg)

	if called {
		t.Fatal("got a relay call with an empty --project, want the fetch skipped entirely")
	}
}
