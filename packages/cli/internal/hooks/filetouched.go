package hooks

import "github.com/francogalfre/coop/packages/cli/internal/redact"

// fileTouchedFields derives an optional file.touched event from a
// PostToolUse payload. Path comes from tool_input (not tool_response) since
// that's where every observed tool that touches a file puts it; a tool with
// no file_path (e.g. Bash) yields no event.
func fileTouchedFields(toolName string, payload map[string]any, red *redact.Redactor, streamDiffs bool) (map[string]any, bool, error) {
	mode, ok := fileTouchedMode(toolName)
	if !ok {
		return nil, false, nil
	}

	toolInput, _ := payload["tool_input"].(map[string]any)

	path := stringField(toolInput, "file_path")
	if path == "" {
		return nil, false, nil
	}

	redactedPath, _ := red.Text(path)

	fields := map[string]any{
		"type": "file.touched",
		"mode": mode,
		"path": redactedPath,
	}

	toolResponse, _ := payload["tool_response"].(map[string]any)

	if patchHunks, ok := parseStructuredPatch(toolResponse["structuredPatch"]); ok {
		added, removed := diffCounts(patchHunks)
		fields["added"] = added
		fields["removed"] = removed

		if streamDiffs {
			hunks, err := buildDiffHunks(patchHunks, red)
			if err != nil {
				return nil, false, err
			}

			fields["hunks"] = hunks
		}
	}

	return fields, true, nil
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
