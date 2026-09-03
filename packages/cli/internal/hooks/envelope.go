package hooks

import (
	"encoding/json"
	"fmt"
	"os/user"

	"github.com/francogalfre/coop/packages/cli/internal/capabilities"
	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/redact"
	"github.com/francogalfre/coop/packages/cli/internal/repoid"
)

const textLimit = 8192

func buildEventBody(cfg config.Config, nextSeq func() int, sessionID, hookEvent string, payload map[string]any, red *redact.Redactor) ([][]byte, error) {
	var bodies [][]byte

	emit := func(fields map[string]any) error {
		return emitEnvelope(&bodies, nextSeq, sessionID, fields)
	}

	switch hookEvent {
	case "SessionStart":
		if err := emitSessionStart(cfg, emit, payload, red); err != nil {
			return nil, err
		}

	case "SessionEnd":
		if err := emitSessionEnd(emit, payload); err != nil {
			return nil, err
		}

	case "PreToolUse":
		if err := emitToolCall(emit, payload, red); err != nil {
			return nil, err
		}

	case "PostToolUse":
		if err := emitToolResult(cfg, emit, payload, red); err != nil {
			return nil, err
		}

	case "Stop":
		if err := emitStop(emit, payload, red); err != nil {
			return nil, err
		}

	case "UserPromptSubmit":
		if err := emitHumanPrompt(emit, payload, red); err != nil {
			return nil, err
		}

	default:
		redactedPayload, _ := red.Value(payload)
		if err := emit(map[string]any{"type": "hook." + hookEvent, "raw": redactedPayload}); err != nil {
			return nil, err
		}
	}

	return bodies, nil
}

func emitSessionStart(cfg config.Config, emit func(map[string]any) error, payload map[string]any, red *redact.Redactor) error {
	rawCwd := stringField(payload, "cwd")
	cwd, _ := red.Text(rawCwd)

	fields := map[string]any{
		"type":         "session.start",
		"harness":      "claude-code",
		"cwd":          cwd,
		"owner":        actorFields(cfg),
		"capabilities": capabilities.ForAttach("claude-code"),
	}

	if rawCwd != "" {
		if repo := repoid.Detect(rawCwd); repo != "" {
			fields["repo"] = repo
		}
	}

	return emit(fields)
}

func emitToolCall(emit func(map[string]any) error, payload map[string]any, red *redact.Redactor) error {
	toolName, _ := red.Text(stringField(payload, "tool_name"))
	if toolName == "" {
		return nil
	}

	toolUseID, _ := red.Text(stringField(payload, "tool_use_id"))

	input, err := redactedTextFrom(payload["tool_input"], red)
	if err != nil {
		return err
	}

	return emit(map[string]any{
		"type":        "tool.call",
		"tool_name":   toolName,
		"tool_use_id": toolUseID,
		"input":       input,
	})
}

func emitToolResult(cfg config.Config, emit func(map[string]any) error, payload map[string]any, red *redact.Redactor) error {
	rawToolName := stringField(payload, "tool_name")

	toolName, _ := red.Text(rawToolName)
	if toolName == "" {
		return nil
	}

	toolUseID, _ := red.Text(stringField(payload, "tool_use_id"))

	output, err := redactedTextFrom(payload["tool_response"], red)
	if err != nil {
		return err
	}

	fields := map[string]any{
		"type":        "tool.result",
		"tool_name":   toolName,
		"tool_use_id": toolUseID,
		"output":      output,
	}

	if durationMs, ok := payload["duration_ms"].(float64); ok {
		fields["duration_ms"] = int64(durationMs)
	}

	if err := emit(fields); err != nil {
		return err
	}

	touched, ok, err := fileTouchedFields(rawToolName, payload, red, !cfg.DisableDiffStreaming)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	touched["tool_use_id"] = toolUseID

	return emit(touched)
}

func emitStop(emit func(map[string]any) error, payload map[string]any, red *redact.Redactor) error {
	if err := emit(map[string]any{"type": "agent.turn_end"}); err != nil {
		return err
	}

	text, err := redactedTextFrom(payload["last_assistant_message"], red)
	if err != nil {
		return err
	}

	return emit(map[string]any{"type": "agent.text", "text": text})
}

func emitHumanPrompt(emit func(map[string]any) error, payload map[string]any, red *redact.Redactor) error {
	text, err := redactedTextFrom(payload["prompt"], red)
	if err != nil {
		return err
	}

	return emit(map[string]any{"type": "human.prompt", "text": text})
}

func emitSessionEnd(emit func(map[string]any) error, payload map[string]any) error {
	fields := map[string]any{"type": "session.end"}

	switch reason := stringField(payload, "reason"); reason {
	case "completed", "cancelled", "error":
		fields["reason"] = reason
	default:
		// A live `claude -p` exit reports reason "other" (harnesses.md); map
		// any unrecognized value to "completed" rather than drop it.
		fields["reason"] = "completed"
	}

	return emit(fields)
}

func stringField(fields map[string]any, key string) string {
	if s, ok := fields[key].(string); ok {
		return s
	}

	return ""
}

func redactedTextFrom(v any, red *redact.Redactor) (map[string]any, error) {
	if v == nil {
		return map[string]any{"text": "", "redactions": 0, "truncated": false}, nil
	}

	redactedVal, count := red.Value(v)

	// A plain string (last_assistant_message, prompt) is the text as-is; only
	// a structured value (tool_input, tool_response) needs JSON-stringifying
	// into something readable. Marshaling a string here would double-encode
	// it, wrapping it in quotes and escaping its own newlines.
	text, ok := redactedVal.(string)
	if !ok {
		b, err := json.Marshal(redactedVal)
		if err != nil {
			return nil, fmt.Errorf("hooks: marshal redacted field: %w", err)
		}

		text = string(b)
	}

	truncated := false

	if len(text) > textLimit {
		text = text[:textLimit]
		truncated = true
	}

	return map[string]any{"text": text, "redactions": count, "truncated": truncated}, nil
}

func actorFields(cfg config.Config) map[string]string {
	if cfg.Username != "" || cfg.DisplayName != "" {
		name := cfg.DisplayName
		if name == "" {
			name = cfg.Username
		}
		id := cfg.UserID
		if id == "" {
			id = cfg.Username
		}
		if id == "" {
			id = name
		}
		fields := map[string]string{"id": id, "display_name": name}
		if cfg.AvatarURL != "" {
			fields["avatar_url"] = cfg.AvatarURL
		}
		return fields
	}

	id, name := "local", "local"

	if u, err := user.Current(); err == nil && u.Username != "" {
		id = u.Username
		name = u.Username
	}

	return map[string]string{"id": id, "display_name": name}
}
