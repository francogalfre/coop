package hooks

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

func TestHandleHookDeniesPreToolUseWhileTakeoverActive(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setTakeover(true, "Alice")

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

	if out["permissionDecision"] != "deny" {
		t.Fatalf("got permissionDecision %v, want deny", out["permissionDecision"])
	}
	if out["permissionDecisionReason"] != "Alice has taken over this session via coop" {
		t.Fatalf("got permissionDecisionReason %v, want the attributed takeover message", out["permissionDecisionReason"])
	}
}

func TestHandleHookNeverDeniesUnsteerableEventsDuringTakeover(t *testing.T) {
	unsteerable := []string{"SessionStart", "SessionEnd", "Stop"}

	for _, event := range unsteerable {
		t.Run(event, func(t *testing.T) {
			relay := newFakeRelay()
			relayServer := relay.server()
			defer relayServer.Close()
			relay.setTakeover(true, "Alice")

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
				t.Fatalf("event %s got hookSpecificOutput during takeover, want none: %v", event, resp)
			}
		})
	}
}

func TestHandleHookDoesNotDenyOtherHarnessesDuringTakeover(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setTakeover(true, "Alice")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, true)
	defer s.Close()

	rec := doHook(t, s, "opencode", "tool.execute.before", `{}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	steer, _ := resp["steer"].(string)
	if steer != "Alice has taken over this session via coop" {
		t.Fatalf("got steer %q, want an advisory takeover notice", steer)
	}
}

func TestHandleHookOnlyNotifiesTakeoverOnceUntilReleased(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setTakeover(true, "Alice")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}
	s := NewServer(cfg, true)
	defer s.Close()

	first := doHook(t, s, "opencode", "tool.execute.before", `{}`)
	var firstResp map[string]any
	_ = json.Unmarshal(first.Body.Bytes(), &firstResp)
	if firstResp["steer"] == nil {
		t.Fatal("expected the first poll to carry the takeover notice")
	}

	second := doHook(t, s, "opencode", "tool.execute.before", `{}`)
	var secondResp map[string]any
	_ = json.Unmarshal(second.Body.Bytes(), &secondResp)
	if secondResp["steer"] != nil {
		t.Fatalf("expected the second poll to stay silent, got %v", secondResp)
	}

	relay.setTakeover(false, "")
	doHook(t, s, "opencode", "tool.execute.before", `{}`)

	relay.setTakeover(true, "Bob")

	third := doHook(t, s, "opencode", "tool.execute.before", `{}`)
	var thirdResp map[string]any
	_ = json.Unmarshal(third.Body.Bytes(), &thirdResp)
	if thirdResp["steer"] != "Bob has taken over this session via coop" {
		t.Fatalf("expected a fresh notice after a new takeover, got %v", thirdResp)
	}
}
