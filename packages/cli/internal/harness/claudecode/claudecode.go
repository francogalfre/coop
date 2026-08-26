package claudecode

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/francogalfre/coop/packages/cli/internal/harness"
)

const (
	Name             = "claude-code"
	settingsDir      = ".claude"
	settingsFileName = "settings.local.json"
	hookTimeoutSecs  = 5
)

var hookEventNames = []string{
	"SessionStart",
	"SessionEnd",
	"UserPromptSubmit",
	"PreToolUse",
	"PostToolUse",
	"Stop",
}

type Adapter struct{}

func (Adapter) Name() string { return Name }

func (Adapter) Binary() string { return "claude" }

func (Adapter) IsFallback() bool { return false }

func (Adapter) Detect(dir string) bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func (Adapter) Install(dir, baseURL string) (harness.Installation, error) {
	path, err := writeSettings(dir, baseURL)
	if err != nil {
		return harness.Installation{}, err
	}

	return harness.Installation{
		Paths:  []string{path},
		Remove: func() error { return removeSettings(dir) },
	}, nil
}

func (Adapter) RemoveAll(dir string) error {
	return removeSettings(dir)
}

func settingsPath(dir string) string {
	return filepath.Join(dir, settingsDir, settingsFileName)
}

func writeSettings(dir, baseURL string) (string, error) {
	path := settingsPath(dir)

	root := readSettingsRoot(path)

	hooksRoot := map[string][]hookGroup{}
	if hooksRaw, ok := root["hooks"]; ok {
		_ = json.Unmarshal(hooksRaw, &hooksRoot)
	}

	for _, event := range hookEventNames {
		hooksRoot[event] = withCoopEntry(hooksRoot[event], event, baseURL)
	}

	hooksJSON, err := json.Marshal(hooksRoot)
	if err != nil {
		return "", fmt.Errorf("harness: marshal hooks: %w", err)
	}

	root["hooks"] = hooksJSON

	if err := writeSettingsRoot(dir, path, root); err != nil {
		return "", err
	}

	return path, nil
}

func readSettingsRoot(path string) map[string]json.RawMessage {
	root := map[string]json.RawMessage{}

	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &root)
	}

	return root
}

func writeSettingsRoot(dir, path string, root map[string]json.RawMessage) error {
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("harness: marshal settings: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, settingsDir), 0o755); err != nil {
		return fmt.Errorf("harness: create %s: %w", filepath.Join(dir, settingsDir), err)
	}

	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("harness: write %s: %w", path, err)
	}

	return nil
}
