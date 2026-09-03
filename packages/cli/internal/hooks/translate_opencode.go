package hooks

import (
	"github.com/francogalfre/coop/packages/cli/internal/capabilities"
	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/redact"
	"github.com/francogalfre/coop/packages/cli/internal/repoid"
)

func translateOpencode(cfg config.Config, nextSeq func() int, sessionID, event string, payload map[string]any, red *redact.Redactor) ([][]byte, error) {
	var bodies [][]byte

	switch event {
	case "session.created":
		rawDir := stringField(payload, "directory")
		dir, _ := red.Text(rawDir)
		if dir == "" {
			return bodies, emitGenericOpencode(&bodies, nextSeq, sessionID, event, payload, red)
		}

		fields := map[string]any{
			"type":         "session.start",
			"harness":      "opencode",
			"cwd":          dir,
			"owner":        actorFields(cfg),
			"capabilities": capabilities.ForAttach("opencode"),
		}

		if repo := repoid.Detect(rawDir); repo != "" {
			fields["repo"] = repo
		}

		return bodies, emitEnvelope(&bodies, nextSeq, sessionID, fields)

	case "session.idle":
		return bodies, emitEnvelope(&bodies, nextSeq, sessionID, map[string]any{"type": "agent.turn_end"})

	case "tool.execute.before":
		return bodies, emitOpencodeToolCall(&bodies, nextSeq, sessionID, payload, red)

	case "tool.execute.after":
		return bodies, emitOpencodeToolResult(&bodies, nextSeq, sessionID, payload, red)

	case "file.edited":
		return bodies, emitOpencodeFileTouched(&bodies, nextSeq, sessionID, payload, red)

	default:
		return bodies, emitGenericOpencode(&bodies, nextSeq, sessionID, event, payload, red)
	}
}

func emitOpencodeToolCall(bodies *[][]byte, nextSeq func() int, sessionID string, payload map[string]any, red *redact.Redactor) error {
	input, _ := payload["input"].(map[string]any)

	toolName, _ := red.Text(stringField(input, "tool"))
	if toolName == "" {
		return emitGenericOpencode(bodies, nextSeq, sessionID, "tool.execute.before", payload, red)
	}

	toolUseID, _ := red.Text(stringField(input, "callID"))

	args, err := redactedTextFrom(payload["args"], red)
	if err != nil {
		return err
	}

	return emitEnvelope(bodies, nextSeq, sessionID, map[string]any{
		"type":        "tool.call",
		"tool_name":   toolName,
		"tool_use_id": toolUseID,
		"input":       args,
	})
}

func emitOpencodeToolResult(bodies *[][]byte, nextSeq func() int, sessionID string, payload map[string]any, red *redact.Redactor) error {
	input, _ := payload["input"].(map[string]any)

	toolName, _ := red.Text(stringField(input, "tool"))
	if toolName == "" {
		return emitGenericOpencode(bodies, nextSeq, sessionID, "tool.execute.after", payload, red)
	}

	toolUseID, _ := red.Text(stringField(input, "callID"))

	output, err := redactedTextFrom(payload["output"], red)
	if err != nil {
		return err
	}

	return emitEnvelope(bodies, nextSeq, sessionID, map[string]any{
		"type":        "tool.result",
		"tool_name":   toolName,
		"tool_use_id": toolUseID,
		"output":      output,
	})
}

// The bus's file.edited carries only {properties: {file}} — no tool call or
// sessionID context (harnesses.md: "the free file-touched signal, no path
// inference needed"), so this is always a write.
func emitOpencodeFileTouched(bodies *[][]byte, nextSeq func() int, sessionID string, payload map[string]any, red *redact.Redactor) error {
	busEvent, _ := payload["event"].(map[string]any)
	properties, _ := busEvent["properties"].(map[string]any)

	path := stringField(properties, "file")
	if path == "" {
		return emitGenericOpencode(bodies, nextSeq, sessionID, "file.edited", payload, red)
	}

	redactedPath, _ := red.Text(path)

	return emitEnvelope(bodies, nextSeq, sessionID, map[string]any{
		"type": "file.touched",
		"mode": "write",
		"path": redactedPath,
	})
}

func emitGenericOpencode(bodies *[][]byte, nextSeq func() int, sessionID, event string, payload map[string]any, red *redact.Redactor) error {
	redactedPayload, _ := red.Value(payload)

	return emitEnvelope(bodies, nextSeq, sessionID, map[string]any{
		"type": "hook.opencode." + event,
		"raw":  redactedPayload,
	})
}
