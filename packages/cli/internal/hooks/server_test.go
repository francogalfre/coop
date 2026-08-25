package hooks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

type fakeRelay struct {
	mu        sync.Mutex
	ingested  []map[string]any
	steerFrom string
	steerText string
	steerPend bool
}

func newFakeRelay() *fakeRelay {
	return &fakeRelay{}
}

func (f *fakeRelay) setSteer(from, text string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.steerFrom, f.steerText, f.steerPend = from, text, true
}

func (f *fakeRelay) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.ingested)
}

func (f *fakeRelay) waitForIngested(t *testing.T, n int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if f.count() >= n {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %d ingested events, got %d", n, f.count())
}

func (f *fakeRelay) server() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/events", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)

		f.mu.Lock()
		f.ingested = append(f.ingested, body)
		f.mu.Unlock()

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	})

	mux.HandleFunc("GET /v1/sessions/{id}/steer", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()

		if !f.steerPend {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		f.steerPend = false
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"from": f.steerFrom, "text": f.steerText})
	})

	return httptest.NewServer(mux)
}

func doHook(t *testing.T, srv *Server, harnessName, event, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/hook/"+harnessName+"/"+event, strings.NewReader(body))
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	return rec
}

func TestHandleHookForwardsRedactedEventToRelay(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, true)
	defer s.Close()

	body := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_use_id":"toolu_1",` +
		`"tool_input":{"command":"curl -H 'Authorization: Bearer sk-abcdefghijklmnopqrstuv' https://api.example.com"}}`

	rec := doHook(t, s, "claude-code", "PreToolUse", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	relay.waitForIngested(t, 1)

	relay.mu.Lock()
	raw, _ := json.Marshal(relay.ingested[0])
	relay.mu.Unlock()

	if strings.Contains(string(raw), "sk-abcdefghijklmnopqrstuv") {
		t.Fatalf("secret reached the relay: %s", raw)
	}
	if !strings.Contains(string(raw), "[redacted]") {
		t.Fatalf("redaction marker missing from ingested event: %s", raw)
	}
}

func TestHandleHookReturnsSteerOnSteerableEvent(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setSteer("Alice", "try the other branch")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, true)
	defer s.Close()

	rec := doHook(t, s, "claude-code", "PreToolUse", `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	out, ok := resp["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("got response %v, want hookSpecificOutput", resp)
	}

	ctx, _ := out["additionalContext"].(string)
	if ctx != "[Alice via coop] try the other branch" {
		t.Fatalf("got additionalContext %q, want the attributed form", ctx)
	}
}

func TestHandleHookSkipsSteerWhenServerNotOwningSteering(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setSteer("Alice", "try the other branch")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, false)
	defer s.Close()

	rec := doHook(t, s, "claude-code", "PreToolUse", `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if _, ok := resp["hookSpecificOutput"]; ok {
		t.Fatalf("steer:false server returned hookSpecificOutput, want none so the pty poller owns steering: %v", resp)
	}
}

func TestHandleHookNeverAttachesSteerOnUnsteerableEvents(t *testing.T) {
	unsteerable := []string{"SessionStart", "SessionEnd", "Stop"}

	for _, event := range unsteerable {
		t.Run(event, func(t *testing.T) {
			relay := newFakeRelay()
			relayServer := relay.server()
			defer relayServer.Close()
			relay.setSteer("Alice", "do something else")

			cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
			s := NewServer(cfg, true)
			defer s.Close()

			rec := doHook(t, s, "claude-code", event, `{"hook_event_name":"`+event+`","cwd":"/tmp"}`)
			if rec.Code != http.StatusOK {
				t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
			}

			var resp map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if _, ok := resp["hookSpecificOutput"]; ok {
				t.Fatalf("event %s got hookSpecificOutput in response, want none: %v", event, resp)
			}
		})
	}
}

func TestHandleHookMalformedBodyRespondsEmpty(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, true)
	defer s.Close()

	rec := doHook(t, s, "claude-code", "PreToolUse", `{not json`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	if strings.TrimSpace(rec.Body.String()) != "{}" {
		t.Fatalf("got body %q, want {}", rec.Body.String())
	}
}

func TestHandleHookIncrementsSeqAcrossEvents(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, true)
	defer s.Close()

	doHook(t, s, "claude-code", "PreToolUse", `{"hook_event_name":"PreToolUse","tool_name":"Bash"}`)
	doHook(t, s, "claude-code", "PostToolUse", `{"hook_event_name":"PostToolUse","tool_name":"Bash"}`)

	relay.waitForIngested(t, 2)

	relay.mu.Lock()
	defer relay.mu.Unlock()

	if relay.ingested[0]["seq"] == relay.ingested[1]["seq"] {
		t.Fatalf("seq did not increment: %v == %v", relay.ingested[0]["seq"], relay.ingested[1]["seq"])
	}
}
