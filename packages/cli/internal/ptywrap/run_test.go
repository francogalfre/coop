package ptywrap

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

type postedEvent struct {
	body    map[string]any
	auth    string
	project string
}

type fakeSteerRelay struct {
	mu             sync.Mutex
	from           string
	text           string
	pending        bool
	takeoverActive bool
	takeoverBy     string
	events         []postedEvent
}

func (f *fakeSteerRelay) eventsSnapshot() []postedEvent {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]postedEvent(nil), f.events...)
}

func (f *fakeSteerRelay) set(from, text string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.from, f.text, f.pending = from, text, true
}

func (f *fakeSteerRelay) setTakeover(active bool, by string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.takeoverActive, f.takeoverBy = active, by
}

func (f *fakeSteerRelay) server() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/sessions/{id}/steer", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()

		resp := map[string]any{
			"has_message": f.pending,
			"from":        f.from,
			"text":        f.text,
			"takeover":    map[string]any{"active": f.takeoverActive, "by": f.takeoverBy},
		}
		f.pending = false

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("POST /v1/events", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var decoded map[string]any
		_ = json.Unmarshal(body, &decoded)

		f.mu.Lock()
		f.events = append(f.events, postedEvent{
			body:    decoded,
			auth:    r.Header.Get("Authorization"),
			project: r.Header.Get("X-Coop-Project"),
		})
		f.mu.Unlock()

		w.WriteHeader(http.StatusAccepted)
	})

	return httptest.NewServer(mux)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	orig := os.Stdout
	os.Stdout = w

	outCh := make(chan string, 1)
	go func() {
		buf, _ := io.ReadAll(r)
		outCh <- string(buf)
	}()

	fn()

	os.Stdout = orig
	_ = w.Close()

	return <-outCh
}

func TestRunPassesThroughChildOutput(t *testing.T) {
	relay := &fakeSteerRelay{}
	relayServer := relay.server()
	defer relayServer.Close()

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	output := captureStdout(t, func() {
		_ = Run(ctx, cfg, "other", "echo", []string{"hello from pty"})
	})

	if !strings.Contains(output, "hello from pty") {
		t.Fatalf("output = %q, want it to contain the child's output", output)
	}
}

func TestRunInjectsSteerIntoChildStdin(t *testing.T) {
	relay := &fakeSteerRelay{}
	relayServer := relay.server()
	defer relayServer.Close()
	relay.set("Alice", "try the other branch")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	output := captureStdout(t, func() {
		_ = Run(ctx, cfg, "other", "cat", nil)
	})

	want := "[Alice via coop] try the other branch"
	if !strings.Contains(output, want) {
		t.Fatalf("output = %q, want it to contain %q", output, want)
	}
}

func TestRunInjectsTakeoverNoticeIntoChildStdin(t *testing.T) {
	relay := &fakeSteerRelay{}
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setTakeover(true, "Alice")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	output := captureStdout(t, func() {
		_ = Run(ctx, cfg, "other", "cat", nil)
	})

	want := "Alice has taken over this session"
	if !strings.Contains(output, want) {
		t.Fatalf("output = %q, want it to contain %q", output, want)
	}
}

func TestRunPostsSessionStartAndEnd(t *testing.T) {
	relay := &fakeSteerRelay{}
	relayServer := relay.server()
	defer relayServer.Close()

	cfg := config.Config{
		RelayURL:      relayServer.URL,
		SessionID:     "sess-a",
		Project:       "proj-a",
		CLICredential: "tok-a",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	captureStdout(t, func() {
		_ = Run(ctx, cfg, "other", "echo", []string{"hi"})
	})

	events := relay.eventsSnapshot()
	if len(events) != 2 {
		t.Fatalf("got %d posted events, want 2 (session.start, session.end): %+v", len(events), events)
	}

	start := events[0]
	if start.body["type"] != "session.start" {
		t.Errorf("got type %v, want session.start", start.body["type"])
	}
	if start.body["harness"] != "other" {
		t.Errorf("got harness %v, want other", start.body["harness"])
	}
	if start.auth != "Bearer tok-a" {
		t.Errorf("got Authorization %q, want Bearer tok-a", start.auth)
	}
	if start.project != "proj-a" {
		t.Errorf("got X-Coop-Project %q, want proj-a", start.project)
	}

	end := events[1]
	if end.body["type"] != "session.end" {
		t.Errorf("got type %v, want session.end", end.body["type"])
	}
	if end.auth != "Bearer tok-a" {
		t.Errorf("got Authorization %q, want Bearer tok-a", end.auth)
	}
}
