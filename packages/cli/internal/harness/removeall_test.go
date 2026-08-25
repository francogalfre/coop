package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveAllRemovesEveryAdaptersTraces(t *testing.T) {
	dir := t.TempDir()

	if _, err := writeClaudeSettings(dir, "http://127.0.0.1:12345"); err != nil {
		t.Fatalf("writeClaudeSettings() error = %v", err)
	}
	if _, err := (opencodeAdapter{}).Install(dir, "http://127.0.0.1:12345"); err != nil {
		t.Fatalf("opencodeAdapter.Install() error = %v", err)
	}
	if _, err := (piAdapter{}).Install(dir, "http://127.0.0.1:12345"); err != nil {
		t.Fatalf("piAdapter.Install() error = %v", err)
	}

	if err := RemoveAll(dir); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}

	for _, p := range []string{
		claudeSettingsPath(dir),
		filepath.Join(dir, ".opencode", "plugin", "coop.js"),
		filepath.Join(dir, ".pi", "extensions", "coop.ts"),
	} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s still exists after RemoveAll(), err = %v", p, err)
		}
	}
}

func TestRemoveAllNoopWhenNothingInstalled(t *testing.T) {
	dir := t.TempDir()

	if err := RemoveAll(dir); err != nil {
		t.Fatalf("RemoveAll() error = %v, want nil for a directory with nothing installed", err)
	}
}
