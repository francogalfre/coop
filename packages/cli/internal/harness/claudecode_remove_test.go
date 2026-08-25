package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveClaudeSettingsDeletesFileWhenOnlyCoopEntriesExisted(t *testing.T) {
	dir := t.TempDir()

	path, err := writeClaudeSettings(dir, "http://127.0.0.1:12345")
	if err != nil {
		t.Fatalf("writeClaudeSettings() error = %v", err)
	}

	if err := removeClaudeSettings(dir); err != nil {
		t.Fatalf("removeClaudeSettings() error = %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("settings file still exists after removeClaudeSettings, err = %v", err)
	}
}

func TestRemoveClaudeSettingsPreservesForeignEntriesAndOtherKeys(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, settingsDir, settingsFileName)

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}

	existing := `{"permissions":{"allow":["Bash(ls:*)"]},` +
		`"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo other-tool"}]}]}}`
	if err := os.WriteFile(settingsPath, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := writeClaudeSettings(dir, "http://127.0.0.1:12345"); err != nil {
		t.Fatalf("writeClaudeSettings() error = %v", err)
	}

	if err := removeClaudeSettings(dir); err != nil {
		t.Fatalf("removeClaudeSettings() error = %v", err)
	}

	root := readSettings(t, settingsPath)

	if _, ok := root["permissions"]; !ok {
		t.Fatal("permissions key was dropped by removeClaudeSettings")
	}

	var hooksRoot map[string][]hookGroup
	if err := json.Unmarshal(root["hooks"], &hooksRoot); err != nil {
		t.Fatalf("unmarshal hooks: %v", err)
	}

	groups := hooksRoot["PreToolUse"]
	if len(groups) != 1 {
		t.Fatalf("got %d groups for PreToolUse after remove, want 1 (foreign tool's only)", len(groups))
	}

	for _, g := range groups {
		for _, e := range g.Hooks {
			if isCoopEntry(e) {
				t.Error("coop entry survived removeClaudeSettings")
			}
		}
	}

	var command string
	if err := json.Unmarshal(groups[0].Hooks[0].extra["command"], &command); err != nil || command != "echo other-tool" {
		t.Fatalf("foreign command field lost: %v", groups[0].Hooks[0].extra["command"])
	}
}

func TestRemoveClaudeSettingsNoopWhenFileAbsent(t *testing.T) {
	dir := t.TempDir()

	if err := removeClaudeSettings(dir); err != nil {
		t.Fatalf("removeClaudeSettings() error = %v, want nil for absent file", err)
	}
}

func TestRemoveClaudeSettingsIsInverseOfWriteEventKeys(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, settingsDir, settingsFileName)

	if _, err := writeClaudeSettings(dir, "http://127.0.0.1:12345"); err != nil {
		t.Fatalf("writeClaudeSettings() error = %v", err)
	}

	if err := removeClaudeSettings(dir); err != nil {
		t.Fatalf("removeClaudeSettings() error = %v", err)
	}

	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Fatalf("expected settings file removed, err = %v", err)
	}
}
