package hooks

import (
	"strings"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

func editPayload(structuredPatch any) map[string]any {
	return map[string]any{
		"hook_event_name": "PostToolUse",
		"tool_name":       "Edit",
		"tool_use_id":     "toolu_1",
		"duration_ms":     float64(42),
		"tool_input":      map[string]any{"file_path": "/repo/main.go"},
		"tool_response": map[string]any{
			"filePath":        "/repo/main.go",
			"structuredPatch": structuredPatch,
		},
	}
}

func onePatchHunk(lines ...string) []any {
	return []any{
		map[string]any{
			"oldStart": float64(1),
			"oldLines": float64(len(lines)),
			"newStart": float64(1),
			"newLines": float64(len(lines)),
			"lines":    toAnySlice(lines),
		},
	}
}

func toAnySlice(lines []string) []any {
	out := make([]any, len(lines))
	for i, l := range lines {
		out[i] = l
	}
	return out
}

func TestBuildEventBodyPostToolUseEditCarriesDurationMs(t *testing.T) {
	payload := editPayload(onePatchHunk(" line one", "-line two", "+line TWO edited", " line three"))

	bodies, err := buildEventBody(config.Config{}, seqFrom(0), "sess-a", "PostToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}

	result := decodeBody(t, bodies[0])
	if result["type"] != "tool.result" {
		t.Fatalf("got type %v, want tool.result", result["type"])
	}
	if result["duration_ms"] != float64(42) {
		t.Errorf("got duration_ms %v, want 42", result["duration_ms"])
	}
}

func TestBuildEventBodyPostToolUseEditEmitsHunksWithAddedRemoved(t *testing.T) {
	payload := editPayload(onePatchHunk(" line one", "-line two", "+line TWO edited", " line three"))

	bodies, err := buildEventBody(config.Config{}, seqFrom(0), "sess-a", "PostToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}

	types := eventTypes(t, bodies)
	if len(types) != 2 || types[1] != "file.touched" {
		t.Fatalf("got types %v, want [tool.result file.touched]", types)
	}

	touched := decodeBody(t, bodies[1])
	if touched["added"] != float64(1) {
		t.Errorf("got added %v, want 1", touched["added"])
	}
	if touched["removed"] != float64(1) {
		t.Errorf("got removed %v, want 1", touched["removed"])
	}

	hunks, ok := touched["hunks"].([]any)
	if !ok || len(hunks) != 1 {
		t.Fatalf("got hunks %v, want exactly 1", touched["hunks"])
	}

	hunk := hunks[0].(map[string]any)
	if hunk["old_start"] != float64(1) || hunk["old_lines"] != float64(4) {
		t.Errorf("got old_start=%v old_lines=%v, want 1, 4", hunk["old_start"], hunk["old_lines"])
	}

	lines, ok := hunk["lines"].([]any)
	if !ok || len(lines) != 4 {
		t.Fatalf("got lines %v, want 4", hunk["lines"])
	}

	firstLine := lines[0].(map[string]any)
	if firstLine["text"] != " line one" {
		t.Errorf("got first line %v, want %q", firstLine["text"], " line one")
	}
}

func TestBuildEventBodyPostToolUseEditRedactsSecretInsideHunkLine(t *testing.T) {
	payload := editPayload(onePatchHunk(
		" line one",
		"+const key = \"sk-abcdefghijklmnopqrstuv\"",
		" line three",
	))

	bodies, err := buildEventBody(config.Config{}, seqFrom(0), "sess-a", "PostToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}

	touched := decodeBody(t, bodies[1])
	hunks := touched["hunks"].([]any)
	hunk := hunks[0].(map[string]any)
	lines := hunk["lines"].([]any)

	secretLine := lines[1].(map[string]any)
	text, _ := secretLine["text"].(string)

	if strings.Contains(text, "sk-abcdefghijklmnopqrstuv") {
		t.Fatalf("hunk line still contains the secret: %q", text)
	}
	if !strings.Contains(text, "[redacted]") {
		t.Fatalf("hunk line = %q, want it to contain [redacted]", text)
	}
}

func TestBuildEventBodyPostToolUseEditOmitsHunksWhenDiffStreamingDisabled(t *testing.T) {
	payload := editPayload(onePatchHunk(" line one", "-line two", "+line TWO edited", " line three"))

	cfg := config.Config{DisableDiffStreaming: true}

	bodies, err := buildEventBody(cfg, seqFrom(0), "sess-a", "PostToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}

	touched := decodeBody(t, bodies[1])
	if _, ok := touched["hunks"]; ok {
		t.Fatalf("got hunks %v, want none when diff streaming is disabled", touched["hunks"])
	}
	if touched["added"] != float64(1) || touched["removed"] != float64(1) {
		t.Errorf("got added=%v removed=%v, want counts to still be reported", touched["added"], touched["removed"])
	}
}

func TestBuildEventBodyPostToolUseWriteEmptyStructuredPatchOmitsHunks(t *testing.T) {
	payload := editPayload([]any{})
	payload["tool_name"] = "Write"

	bodies, err := buildEventBody(config.Config{}, seqFrom(0), "sess-a", "PostToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}

	touched := decodeBody(t, bodies[1])
	if _, ok := touched["hunks"]; ok {
		t.Fatalf("got hunks %v, want none for an empty structuredPatch (a full file create)", touched["hunks"])
	}
	if _, ok := touched["added"]; ok {
		t.Fatalf("got added %v, want none: an empty structuredPatch carries no diff data", touched["added"])
	}
}

func TestBuildDiffHunksCapsHunksAndLinesPerHunk(t *testing.T) {
	manyHunks := make([]structuredPatchHunk, 25)
	for i := range manyHunks {
		manyHunks[i] = structuredPatchHunk{oldStart: i, oldLines: 1, newStart: i, newLines: 1, lines: []string{" x"}}
	}

	out, err := buildDiffHunks(manyHunks, redact.New())
	if err != nil {
		t.Fatalf("buildDiffHunks() error = %v", err)
	}
	if len(out) != maxDiffHunks {
		t.Fatalf("got %d hunks, want capped at %d", len(out), maxDiffHunks)
	}

	longLines := make([]string, 450)
	for i := range longLines {
		longLines[i] = " x"
	}

	out, err = buildDiffHunks([]structuredPatchHunk{{lines: longLines}}, redact.New())
	if err != nil {
		t.Fatalf("buildDiffHunks() error = %v", err)
	}

	lines := out[0]["lines"].([]map[string]any)
	if len(lines) != maxDiffLines {
		t.Fatalf("got %d lines, want capped at %d", len(lines), maxDiffLines)
	}
}

func TestParseStructuredPatchReturnsFalseForMissingOrEmpty(t *testing.T) {
	if _, ok := parseStructuredPatch(nil); ok {
		t.Error("parseStructuredPatch(nil) = true, want false")
	}
	if _, ok := parseStructuredPatch([]any{}); ok {
		t.Error("parseStructuredPatch([]) = true, want false")
	}
	if _, ok := parseStructuredPatch("not an array"); ok {
		t.Error("parseStructuredPatch(string) = true, want false")
	}
}
