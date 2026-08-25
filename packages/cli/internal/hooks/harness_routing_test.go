package hooks

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

func TestHandleHookOpencodeSessionCreatedTranslatesToSessionStart(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, true)
	defer s.Close()

	rec := doHook(t, s, "opencode", "session.created", `{"directory":"/repo","worktree":"/repo"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	relay.waitForIngested(t, 1)

	relay.mu.Lock()
	got := relay.ingested[0]
	relay.mu.Unlock()

	if got["type"] != "session.start" || got["harness"] != "opencode" {
		t.Fatalf("got event %v, want session.start/opencode", got)
	}
}

func TestHandleHookPiToolCallStillReturnsSteerAsJSONField(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setSteer("Alice", "look at the other file")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, true)
	defer s.Close()

	rec := doHook(t, s, "pi", "tool_call",
		`{"toolName":"read","toolCallId":"tc_1","input":{"path":"/repo/main.go"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	steer, _ := resp["steer"].(string)
	if steer != "[Alice via coop] look at the other file" {
		t.Fatalf("got steer %q, want the attributed form (pi/opencode use a JSON field, not additionalContext)", steer)
	}

	relay.waitForIngested(t, 2)

	relay.mu.Lock()
	defer relay.mu.Unlock()

	types := []string{}
	for _, e := range relay.ingested {
		types = append(types, e["type"].(string))
	}

	if len(types) != 2 || types[0] != "tool.call" || types[1] != "file.touched" {
		t.Fatalf("got event types %v, want [tool.call file.touched]", types)
	}
}
