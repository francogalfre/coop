package hooks

import "github.com/francogalfre/coop/packages/cli/internal/redact"

// fileTouchedFields derives an optional file.touched event from a
// PostToolUse payload. Path comes from tool_input (not tool_response) since
// that's where every observed tool that touches a file puts it; a tool with
// no file_path (e.g. Bash) yields no event.
func fileTouchedFields(toolName string, payload map[string]any, red *redact.Redactor) (map[string]any, bool) {
	mode, ok := fileTouchedMode(toolName)
	if !ok {
		return nil, false
	}

	toolInput, _ := payload["tool_input"].(map[string]any)

	path := stringField(toolInput, "file_path")
	if path == "" {
		return nil, false
	}

	redactedPath, _ := red.Text(path)

	return map[string]any{
		"type": "file.touched",
		"mode": mode,
		"path": redactedPath,
	}, true
}

func fileTouchedMode(toolName string) (string, bool) {
	switch toolName {
	case "Read", "Grep", "Glob":
		return "read", true
	case "Write", "Edit", "NotebookEdit":
		return "write", true
	default:
		return "", false
	}
}
