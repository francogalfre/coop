package hooks

import (
	"strings"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

func translatePiOne(t *testing.T, seq int, sessionID, event string, payload map[string]any, red *redact.Redactor) map[string]any {
	t.Helper()

	bodies, err := translatePi(config.Config{}, seqFrom(seq), sessionID, event, payload, red)
	if err != nil {
		t.Fatalf("translatePi() error = %v", err)
	}
	if len(bodies) != 1 {
		t.Fatalf("got %d events, want exactly 1: %v", len(bodies), bodies)
	}

	return decodeBody(t, bodies[0])
}

func TestTranslatePiToolResultEmitsToolResult(t *testing.T) {
	payload := map[string]any{
		"toolName":   "bash",
		"toolCallId": "tc_1",
		"input":      map[string]any{"command": "ls"},
		"content": []any{
			map[string]any{"type": "text", "text": "total 0\nsecret sk-abcdefghijklmnopqrstuv"},
		},
		"isError": false,
	}

	out := translatePiOne(t, 0, "sess-a", "tool_result", payload, redact.New())

	if out["type"] != "tool.result" {
		t.Errorf("got type %v, want tool.result", out["type"])
	}
	if out["tool_name"] != "bash" {
		t.Errorf("got tool_name %v, want bash", out["tool_name"])
	}
	if out["tool_use_id"] != "tc_1" {
		t.Errorf("got tool_use_id %v, want tc_1", out["tool_use_id"])
	}

	output, ok := out["output"].(map[string]any)
	if !ok {
		t.Fatalf("got output %v, want an object", out["output"])
	}

	text, _ := output["text"].(string)
	if strings.Contains(text, "sk-abcdefghijklmnopqrstuv") {
		t.Errorf("output.text still contains the secret: %s", text)
	}
	if !strings.Contains(text, "[redacted]") {
		t.Errorf("output.text = %q, want it to contain [redacted]", text)
	}
}

func TestTranslatePiToolResultEmptyToolNameFallsBackToGeneric(t *testing.T) {
	payload := map[string]any{"content": []any{}}

	out := translatePiOne(t, 1, "sess-a", "tool_result", payload, redact.New())

	if out["type"] != "hook.pi.tool_result" {
		t.Errorf("got type %v, want hook.pi.tool_result", out["type"])
	}
}

func TestTranslatePiAgentEndEmitsAgentTurnEnd(t *testing.T) {
	payload := map[string]any{
		"messages": []any{map[string]any{"role": "assistant", "content": "done"}},
	}

	out := translatePiOne(t, 2, "sess-a", "agent_end", payload, redact.New())

	if out["type"] != "agent.turn_end" {
		t.Errorf("got type %v, want agent.turn_end", out["type"])
	}
}
