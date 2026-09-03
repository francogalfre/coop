package hooks

import (
	"github.com/francogalfre/coop/packages/cli/internal/capabilities"
	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/redact"
	"github.com/francogalfre/coop/packages/cli/internal/repoid"
)

func translatePi(cfg config.Config, nextSeq func() int, sessionID, event string, payload map[string]any, red *redact.Redactor) ([][]byte, error) {
	var bodies [][]byte

	switch event {
	case "session_start":
		if err := emitPiSessionStart(cfg, &bodies, nextSeq, sessionID, payload, red); err != nil {
			return nil, err
		}

	case "tool_call":
		if err := emitPiToolCall(&bodies, nextSeq, sessionID, payload, red); err != nil {
			return nil, err
		}

	case "tool_result":
		if err := emitPiToolResult(&bodies, nextSeq, sessionID, payload, red); err != nil {
			return nil, err
		}

	case "turn_end", "agent_end":
		if err := emitEnvelope(&bodies, nextSeq, sessionID, map[string]any{"type": "agent.turn_end"}); err != nil {
			return nil, err
		}

	case "session_shutdown":
		if err := emitEnvelope(&bodies, nextSeq, sessionID, map[string]any{"type": "session.end"}); err != nil {
			return nil, err
		}

	default:
		if err := emitGenericPi(&bodies, nextSeq, sessionID, event, payload, red); err != nil {
			return nil, err
		}
	}

	return bodies, nil
}

func emitPiSessionStart(cfg config.Config, bodies *[][]byte, nextSeq func() int, sessionID string, payload map[string]any, red *redact.Redactor) error {
	rawCwd := stringField(payload, "cwd")
	cwd, _ := red.Text(rawCwd)
	if cwd == "" {
		return emitGenericPi(bodies, nextSeq, sessionID, "session_start", payload, red)
	}

	fields := map[string]any{
		"type":         "session.start",
		"harness":      "pi",
		"cwd":          cwd,
		"owner":        actorFields(cfg),
		"capabilities": capabilities.ForAttach("pi"),
	}

	if repo := repoid.Detect(rawCwd); repo != "" {
		fields["repo"] = repo
	}

	return emitEnvelope(bodies, nextSeq, sessionID, fields)
}

func emitPiToolCall(bodies *[][]byte, nextSeq func() int, sessionID string, payload map[string]any, red *redact.Redactor) error {
	toolName, _ := red.Text(stringField(payload, "toolName"))
	if toolName == "" {
		return emitGenericPi(bodies, nextSeq, sessionID, "tool_call", payload, red)
	}

	toolCallID, _ := red.Text(stringField(payload, "toolCallId"))

	input, err := redactedTextFrom(payload["input"], red)
	if err != nil {
		return err
	}

	if err := emitEnvelope(bodies, nextSeq, sessionID, map[string]any{
		"type":        "tool.call",
		"tool_name":   toolName,
		"tool_use_id": toolCallID,
		"input":       input,
	}); err != nil {
		return err
	}

	return emitPiFileTouched(bodies, nextSeq, sessionID, stringField(payload, "toolName"), payload, toolCallID, red)
}

func emitPiFileTouched(bodies *[][]byte, nextSeq func() int, sessionID, toolName string, payload map[string]any, toolCallID string, red *redact.Redactor) error {
	mode, ok := piFileTouchedMode(toolName)
	if !ok {
		return nil
	}

	inputMap, _ := payload["input"].(map[string]any)

	path := stringField(inputMap, "path")
	if path == "" {
		return nil
	}

	redactedPath, _ := red.Text(path)

	return emitEnvelope(bodies, nextSeq, sessionID, map[string]any{
		"type":        "file.touched",
		"mode":        mode,
		"path":        redactedPath,
		"tool_use_id": toolCallID,
	})
}

func emitPiToolResult(bodies *[][]byte, nextSeq func() int, sessionID string, payload map[string]any, red *redact.Redactor) error {
	toolName, _ := red.Text(stringField(payload, "toolName"))
	if toolName == "" {
		return emitGenericPi(bodies, nextSeq, sessionID, "tool_result", payload, red)
	}

	toolCallID, _ := red.Text(stringField(payload, "toolCallId"))

	output, err := redactedTextFrom(payload["content"], red)
	if err != nil {
		return err
	}

	return emitEnvelope(bodies, nextSeq, sessionID, map[string]any{
		"type":        "tool.result",
		"tool_name":   toolName,
		"tool_use_id": toolCallID,
		"output":      output,
	})
}

func piFileTouchedMode(toolName string) (string, bool) {
	switch toolName {
	case "read":
		return "read", true
	case "edit", "write":
		return "write", true
	default:
		return "", false
	}
}

func emitGenericPi(bodies *[][]byte, nextSeq func() int, sessionID, event string, payload map[string]any, red *redact.Redactor) error {
	redactedPayload, _ := red.Value(payload)

	return emitEnvelope(bodies, nextSeq, sessionID, map[string]any{
		"type": "hook.pi." + event,
		"raw":  redactedPayload,
	})
}
