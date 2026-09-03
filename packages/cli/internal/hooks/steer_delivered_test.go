package hooks

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

func TestHandleHookDropsCommandInAttachModeNoPty(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setSteerReceipt("Alice", "/model sonnet", "steer-1", "command")

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

	if _, ok := resp["hookSpecificOutput"]; ok {
		t.Fatalf("got hookSpecificOutput %v, want none: a command must never be delivered from attach mode (no pty)", resp)
	}

	relay.mu.Lock()
	defer relay.mu.Unlock()

	for _, e := range relay.ingested {
		if e["type"] == "steer.delivered" {
			t.Fatalf("got a steer.delivered event for a dropped command: %v", e)
		}
	}
}

func TestHandleHookEmitsSteerDeliveredWithHookEventWhenIDPresent(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setSteerReceipt("Alice", "try the other branch", "steer-1", "steer")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, true)
	defer s.Close()

	rec := doHook(t, s, "claude-code", "PreToolUse", `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	relay.waitForIngested(t, 2)

	relay.mu.Lock()
	defer relay.mu.Unlock()

	var delivered map[string]any
	for _, e := range relay.ingested {
		if e["type"] == "steer.delivered" {
			delivered = e
		}
	}

	if delivered == nil {
		t.Fatalf("no steer.delivered event ingested: %v", relay.ingested)
	}
	if delivered["steer_id"] != "steer-1" {
		t.Errorf("got steer_id %v, want steer-1", delivered["steer_id"])
	}
	if delivered["hook_event"] != "PreToolUse" {
		t.Errorf("got hook_event %v, want PreToolUse", delivered["hook_event"])
	}
}

func TestHandleHookSkipsSteerDeliveredWhenIDAbsent(t *testing.T) {
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
	if _, ok := resp["hookSpecificOutput"].(map[string]any); !ok {
		t.Fatalf("expected steer text still delivered even without an id, got %v", resp)
	}

	relay.mu.Lock()
	defer relay.mu.Unlock()

	for _, e := range relay.ingested {
		if e["type"] == "steer.delivered" {
			t.Fatalf("got a steer.delivered event with no id in the relay response: %v", e)
		}
	}
}
