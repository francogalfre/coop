package hooks

import (
	"strings"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

func translateOpencodeOne(t *testing.T, seq int, sessionID, event string, payload map[string]any, red *redact.Redactor) map[string]any {
	t.Helper()

	bodies, err := translateOpencode(config.Config{}, seqFrom(seq), sessionID, event, payload, red)
	if err != nil {
		t.Fatalf("translateOpencode() error = %v", err)
	}
	if len(bodies) != 1 {
		t.Fatalf("got %d events, want exactly 1: %v", len(bodies), bodies)
	}

	return decodeBody(t, bodies[0])
}

func TestTranslateOpencodeToolExecuteBeforeEmitsToolCall(t *testing.T) {
	payload := map[string]any{
		"directory": "/repo",
		"input":     map[string]any{"tool": "bash", "sessionID": "ses_1", "callID": "call_1"},
		"args": map[string]any{
			"command": "curl -H 'Authorization: Bearer sk-abcdefghijklmnopqrstuv' https://api.example.com",
		},
	}

	out := translateOpencodeOne(t, 0, "sess-a", "tool.execute.before", payload, redact.New())

	if out["type"] != "tool.call" {
		t.Errorf("got type %v, want tool.call", out["type"])
	}
	if out["tool_name"] != "bash" {
		t.Errorf("got tool_name %v, want bash", out["tool_name"])
	}
	if out["tool_use_id"] != "call_1" {
		t.Errorf("got tool_use_id %v, want call_1", out["tool_use_id"])
	}

	input, ok := out["input"].(map[string]any)
	if !ok {
		t.Fatalf("got input %v, want an object", out["input"])
	}

	text, _ := input["text"].(string)
	if strings.Contains(text, "sk-abcdefghijklmnopqrstuv") {
		t.Errorf("input.text still contains the secret: %s", text)
	}
	if !strings.Contains(text, "[redacted]") {
		t.Errorf("input.text = %q, want it to contain [redacted]", text)
	}
}

func TestTranslateOpencodeToolExecuteBeforeEmptyToolFallsBackToGeneric(t *testing.T) {
	payload := map[string]any{"input": map[string]any{}, "args": map[string]any{}}

	out := translateOpencodeOne(t, 1, "sess-a", "tool.execute.before", payload, redact.New())

	if out["type"] != "hook.opencode.tool.execute.before" {
		t.Errorf("got type %v, want hook.opencode.tool.execute.before", out["type"])
	}
}

func TestTranslateOpencodeToolExecuteAfterEmitsToolResult(t *testing.T) {
	payload := map[string]any{
		"directory": "/repo",
		"input": map[string]any{
			"tool": "bash", "sessionID": "ses_1", "callID": "call_1",
			"args": map[string]any{"command": "ls"},
		},
		"title":    "ls",
		"output":   "total 0\nsecret sk-abcdefghijklmnopqrstuv",
		"metadata": map[string]any{},
	}

	out := translateOpencodeOne(t, 2, "sess-a", "tool.execute.after", payload, redact.New())

	if out["type"] != "tool.result" {
		t.Errorf("got type %v, want tool.result", out["type"])
	}
	if out["tool_name"] != "bash" {
		t.Errorf("got tool_name %v, want bash", out["tool_name"])
	}
	if out["tool_use_id"] != "call_1" {
		t.Errorf("got tool_use_id %v, want call_1", out["tool_use_id"])
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

func TestTranslateOpencodeToolExecuteAfterEmptyToolFallsBackToGeneric(t *testing.T) {
	payload := map[string]any{"input": map[string]any{}, "output": "x"}

	out := translateOpencodeOne(t, 3, "sess-a", "tool.execute.after", payload, redact.New())

	if out["type"] != "hook.opencode.tool.execute.after" {
		t.Errorf("got type %v, want hook.opencode.tool.execute.after", out["type"])
	}
}

func TestTranslateOpencodeFileEditedEmitsFileTouched(t *testing.T) {
	payload := map[string]any{
		"directory": "/repo",
		"worktree":  "/repo",
		"event": map[string]any{
			"type":       "file.edited",
			"properties": map[string]any{"file": "/repo/main.go"},
		},
	}

	out := translateOpencodeOne(t, 4, "sess-a", "file.edited", payload, redact.New())

	if out["type"] != "file.touched" {
		t.Errorf("got type %v, want file.touched", out["type"])
	}
	if out["mode"] != "write" {
		t.Errorf("got mode %v, want write", out["mode"])
	}
	if out["path"] != "/repo/main.go" {
		t.Errorf("got path %v, want /repo/main.go", out["path"])
	}
}

func TestTranslateOpencodeFileEditedMissingPathFallsBackToGeneric(t *testing.T) {
	payload := map[string]any{
		"event": map[string]any{"type": "file.edited", "properties": map[string]any{}},
	}

	out := translateOpencodeOne(t, 5, "sess-a", "file.edited", payload, redact.New())

	if out["type"] != "hook.opencode.file.edited" {
		t.Errorf("got type %v, want hook.opencode.file.edited", out["type"])
	}
}
