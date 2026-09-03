package hooks

import (
	"strings"

	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

const (
	maxDiffHunks = 20
	maxDiffLines = 400
)

type structuredPatchHunk struct {
	oldStart, oldLines, newStart, newLines int
	lines                                  []string
}

// parseStructuredPatch reads Claude Code's PostToolUse tool_response.structuredPatch
// (live-verified 2026-09-03, harnesses.md): an array of {oldStart, oldLines,
// newStart, newLines, lines} objects, each line prefixed " ", "+", or "-".
func parseStructuredPatch(raw any) ([]structuredPatchHunk, bool) {
	rawHunks, ok := raw.([]any)
	if !ok || len(rawHunks) == 0 {
		return nil, false
	}

	hunks := make([]structuredPatchHunk, 0, len(rawHunks))

	for _, rh := range rawHunks {
		m, ok := rh.(map[string]any)
		if !ok {
			continue
		}

		hunks = append(hunks, structuredPatchHunk{
			oldStart: intField(m, "oldStart"),
			oldLines: intField(m, "oldLines"),
			newStart: intField(m, "newStart"),
			newLines: intField(m, "newLines"),
			lines:    stringSlice(m["lines"]),
		})
	}

	if len(hunks) == 0 {
		return nil, false
	}

	return hunks, true
}

func diffCounts(hunks []structuredPatchHunk) (added, removed int) {
	for _, h := range hunks {
		for _, line := range h.lines {
			switch {
			case strings.HasPrefix(line, "+"):
				added++
			case strings.HasPrefix(line, "-"):
				removed++
			}
		}
	}

	return added, removed
}

// buildDiffHunks maps structuredPatch's camelCase onto the protocol's
// diffHunk shape, redacting every line - a diff moves file contents off the
// machine, so this is the highest-risk text in the event (security.md) -
// and capping at the protocol's limits (20 hunks, 400 lines/hunk).
func buildDiffHunks(hunks []structuredPatchHunk, red *redact.Redactor) ([]map[string]any, error) {
	if len(hunks) > maxDiffHunks {
		hunks = hunks[:maxDiffHunks]
	}

	out := make([]map[string]any, 0, len(hunks))

	for _, h := range hunks {
		lines := h.lines
		if len(lines) > maxDiffLines {
			lines = lines[:maxDiffLines]
		}

		redactedLines := make([]map[string]any, 0, len(lines))

		for _, line := range lines {
			redactedLine, err := redactedTextFrom(line, red)
			if err != nil {
				return nil, err
			}

			redactedLines = append(redactedLines, redactedLine)
		}

		out = append(out, map[string]any{
			"old_start": h.oldStart,
			"old_lines": h.oldLines,
			"new_start": h.newStart,
			"new_lines": h.newLines,
			"lines":     redactedLines,
		})
	}

	return out, nil
}

func intField(m map[string]any, key string) int {
	if f, ok := m[key].(float64); ok {
		return int(f)
	}

	return 0
}

func stringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}

	out := make([]string, 0, len(arr))

	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}

	return out
}
