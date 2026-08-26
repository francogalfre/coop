package hooks

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

func seqFrom(start int) func() int {
	n := start
	return func() int {
		v := n
		n++
		return v
	}
}

func decodeBody(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal event body: %v (body=%s)", err, body)
	}

	return out
}

func buildOne(t *testing.T, seq int, sessionID, hookEvent string, payload map[string]any, red *redact.Redactor) map[string]any {
	t.Helper()

	bodies, err := buildEventBody(seqFrom(seq), sessionID, hookEvent, payload, red)
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}
	if len(bodies) != 1 {
		t.Fatalf("got %d events, want exactly 1: %v", len(bodies), bodies)
	}

	return decodeBody(t, bodies[0])
}

func eventTypes(t *testing.T, bodies [][]byte) []string {
	t.Helper()

	types := make([]string, len(bodies))
	for i, b := range bodies {
		types[i], _ = decodeBody(t, b)["type"].(string)
	}

	return types
}

func TestBuildEventBodySessionStart(t *testing.T) {
	payload := map[string]any{
		"hook_event_name": "SessionStart",
		"cwd":             "/home/user/project",
		"source":          "startup",
	}

	out := buildOne(t, 0, "sess-a", "SessionStart", payload, redact.New())

	if out["type"] != "session.start" {
		t.Errorf("got type %v, want session.start", out["type"])
	}
	if out["session_id"] != "sess-a" {
		t.Errorf("got session_id %v, want sess-a", out["session_id"])
	}
	if out["cwd"] != "/home/user/project" {
		t.Errorf("got cwd %v, want /home/user/project", out["cwd"])
	}

	owner, ok := out["owner"].(map[string]any)
	if !ok || owner["display_name"] == "" {
		t.Errorf("got owner %v, want a non-empty display_name", out["owner"])
	}
}

func TestBuildEventBodyPreToolUseRedactsInput(t *testing.T) {
	payload := map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_use_id":     "toolu_123",
		"tool_input": map[string]any{
			"command": "curl -H 'Authorization: Bearer sk-abcdefghijklmnopqrstuv' https://api.example.com",
		},
	}

	out := buildOne(t, 1, "sess-a", "PreToolUse", payload, redact.New())

	if out["type"] != "tool.call" {
		t.Errorf("got type %v, want tool.call", out["type"])
	}
	if out["tool_name"] != "Bash" {
		t.Errorf("got tool_name %v, want Bash", out["tool_name"])
	}

	input, ok := out["input"].(map[string]any)
	if !ok {
		t.Fatalf("got input %v (%T), want an object", out["input"], out["input"])
	}

	text, _ := input["text"].(string)
	if strings.Contains(text, "sk-abcdefghijklmnopqrstuv") {
		t.Errorf("input.text still contains the secret: %s", text)
	}
	if !strings.Contains(text, "[redacted]") {
		t.Errorf("input.text = %q, want it to contain [redacted]", text)
	}

	redactions, _ := input["redactions"].(float64)
	if redactions < 1 {
		t.Errorf("got redactions=%v, want at least 1", input["redactions"])
	}
}

func TestBuildEventBodyPreToolUseEmptyToolNameSkipsEvent(t *testing.T) {
	payload := map[string]any{"hook_event_name": "PreToolUse", "tool_input": map[string]any{}}

	bodies, err := buildEventBody(seqFrom(0), "sess-a", "PreToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}
	if len(bodies) != 0 {
		t.Fatalf("got %d events for empty tool_name, want 0", len(bodies))
	}
}

func TestBuildEventBodyPostToolUseBashEmitsOnlyToolResult(t *testing.T) {
	payload := map[string]any{
		"hook_event_name": "PostToolUse",
		"tool_name":       "Bash",
		"tool_use_id":     "toolu_123",
		"tool_input":      map[string]any{"command": "ls"},
		"tool_response":   map[string]any{"stdout": "done", "stderr": ""},
	}

	bodies, err := buildEventBody(seqFrom(2), "sess-a", "PostToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}

	types := eventTypes(t, bodies)
	if len(types) != 1 || types[0] != "tool.result" {
		t.Fatalf("got types %v, want [tool.result]", types)
	}

	out := decodeBody(t, bodies[0])
	output, ok := out["output"].(map[string]any)
	if !ok {
		t.Fatalf("got output %v, want an object", out["output"])
	}
	if !strings.Contains(output["text"].(string), "done") {
		t.Errorf("output.text = %v, want it to contain stdout", output["text"])
	}
}

func TestBuildEventBodyUnmappedEventFallsBackToUnknown(t *testing.T) {
	payload := map[string]any{
		"hook_event_name": "Notification",
		"message":         "use key sk-abcdefghijklmnopqrstuv please",
	}

	out := buildOne(t, 5, "sess-a", "Notification", payload, redact.New())

	if out["type"] != "hook.Notification" {
		t.Errorf("got type %v, want hook.Notification", out["type"])
	}

	raw, ok := out["raw"].(map[string]any)
	if !ok {
		t.Fatalf("got raw %v, want an object", out["raw"])
	}

	message, _ := raw["message"].(string)
	if strings.Contains(message, "sk-abcdefghijklmnopqrstuv") {
		t.Errorf("raw.message still contains the secret: %s", message)
	}
}

func TestBuildEventBodyUserPromptSubmitEmitsHumanPromptRedacted(t *testing.T) {
	payload := map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"prompt":          "use key sk-abcdefghijklmnopqrstuv please",
	}

	out := buildOne(t, 5, "sess-a", "UserPromptSubmit", payload, redact.New())

	if out["type"] != "human.prompt" {
		t.Errorf("got type %v, want human.prompt", out["type"])
	}

	text, ok := out["text"].(map[string]any)
	if !ok {
		t.Fatalf("got text %v, want an object", out["text"])
	}

	prompt, _ := text["text"].(string)
	if strings.Contains(prompt, "sk-abcdefghijklmnopqrstuv") {
		t.Errorf("text.text still contains the secret: %s", prompt)
	}
}

func TestBuildEventBodyEnvelopeFieldsAlwaysPresent(t *testing.T) {
	out := buildOne(t, 7, "sess-z", "hook.custom", map[string]any{}, redact.New())

	if out["v"] != float64(1) {
		t.Errorf("got v=%v, want 1", out["v"])
	}
	if out["session_id"] != "sess-z" {
		t.Errorf("got session_id=%v, want sess-z", out["session_id"])
	}
	if out["seq"] != float64(7) {
		t.Errorf("got seq=%v, want 7", out["seq"])
	}
	if ts, _ := out["ts"].(string); ts == "" {
		t.Error("ts is empty")
	}
}
