package hooks

import (
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

func TestBuildEventBodyPostToolUseReadEmitsFileTouchedRead(t *testing.T) {
	payload := map[string]any{
		"hook_event_name": "PostToolUse",
		"tool_name":       "Read",
		"tool_use_id":     "toolu_5",
		"tool_input":      map[string]any{"file_path": "/repo/main.go"},
		"tool_response":   map[string]any{"content": "package main"},
	}

	bodies, err := buildEventBody(seqFrom(0), "sess-a", "PostToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}

	types := eventTypes(t, bodies)
	if len(types) != 2 || types[0] != "tool.result" || types[1] != "file.touched" {
		t.Fatalf("got types %v, want [tool.result file.touched]", types)
	}

	touched := decodeBody(t, bodies[1])
	if touched["mode"] != "read" {
		t.Errorf("got mode %v, want read", touched["mode"])
	}
	if touched["path"] != "/repo/main.go" {
		t.Errorf("got path %v, want /repo/main.go", touched["path"])
	}

	result := decodeBody(t, bodies[0])
	if touched["seq"] == result["seq"] {
		t.Errorf("file.touched must have its own seq, got both = %v", touched["seq"])
	}
}

func TestBuildEventBodyPostToolUseEditEmitsFileTouchedWrite(t *testing.T) {
	payload := map[string]any{
		"hook_event_name": "PostToolUse",
		"tool_name":       "Edit",
		"tool_use_id":     "toolu_6",
		"tool_input":      map[string]any{"file_path": "/repo/main.go", "old_string": "a", "new_string": "b"},
		"tool_response":   map[string]any{"filePath": "/repo/main.go"},
	}

	bodies, err := buildEventBody(seqFrom(0), "sess-a", "PostToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}

	types := eventTypes(t, bodies)
	if len(types) != 2 || types[0] != "tool.result" || types[1] != "file.touched" {
		t.Fatalf("got types %v, want [tool.result file.touched]", types)
	}

	touched := decodeBody(t, bodies[1])
	if touched["mode"] != "write" {
		t.Errorf("got mode %v, want write", touched["mode"])
	}
}

func TestBuildEventBodySeqIncrementsPerEventNotPerCall(t *testing.T) {
	payload := map[string]any{
		"hook_event_name": "PostToolUse",
		"tool_name":       "Write",
		"tool_input":      map[string]any{"file_path": "/x"},
		"tool_response":   map[string]any{},
	}

	bodies, err := buildEventBody(seqFrom(10), "sess-a", "PostToolUse", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}
	if len(bodies) != 2 {
		t.Fatalf("got %d events, want 2", len(bodies))
	}

	first := decodeBody(t, bodies[0])
	second := decodeBody(t, bodies[1])

	if first["seq"] != float64(10) || second["seq"] != float64(11) {
		t.Fatalf("got seqs %v, %v, want 10, 11", first["seq"], second["seq"])
	}
}
