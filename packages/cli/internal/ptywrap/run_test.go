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

type fakeSteerRelay struct {
	mu      sync.Mutex
	from    string
	text    string
	pending bool
}

func (f *fakeSteerRelay) set(from, text string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.from, f.text, f.pending = from, text, true
}

func (f *fakeSteerRelay) server() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/sessions/{id}/steer", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()

		if !f.pending {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		f.pending = false
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"from": f.from, "text": f.text})
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
		_ = Run(ctx, cfg, "echo", []string{"hello from pty"})
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
		_ = Run(ctx, cfg, "cat", nil)
	})

	want := "[Alice via coop] try the other branch"
	if !strings.Contains(output, want) {
		t.Fatalf("output = %q, want it to contain %q", output, want)
	}
}
