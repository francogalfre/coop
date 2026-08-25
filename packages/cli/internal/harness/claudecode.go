package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	claudeCodeName   = "claude-code"
	settingsDir      = ".claude"
	settingsFileName = "settings.local.json"
	hookTimeoutSecs  = 5
)

var claudeHookEventNames = []string{
	"SessionStart",
	"SessionEnd",
	"UserPromptSubmit",
	"PreToolUse",
	"PostToolUse",
	"Stop",
}

type claudeCodeAdapter struct{}

func (claudeCodeAdapter) Name() string { return claudeCodeName }

func (claudeCodeAdapter) Detect(dir string) bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func (claudeCodeAdapter) Install(dir, baseURL string) (Installation, error) {
	path, err := writeClaudeSettings(dir, baseURL)
	if err != nil {
		return Installation{}, err
	}

	return Installation{
		Paths:  []string{path},
		Remove: func() error { return removeClaudeSettings(dir) },
	}, nil
}

func claudeSettingsPath(dir string) string {
	return filepath.Join(dir, settingsDir, settingsFileName)
}

func writeClaudeSettings(dir, baseURL string) (string, error) {
	path := claudeSettingsPath(dir)

	root := readClaudeSettingsRoot(path)

	hooksRoot := map[string][]hookGroup{}
	if hooksRaw, ok := root["hooks"]; ok {
		_ = json.Unmarshal(hooksRaw, &hooksRoot)
	}

	for _, event := range claudeHookEventNames {
		hooksRoot[event] = withCoopEntry(hooksRoot[event], event, baseURL)
	}

	hooksJSON, err := json.Marshal(hooksRoot)
	if err != nil {
		return "", fmt.Errorf("harness: marshal hooks: %w", err)
	}

	root["hooks"] = hooksJSON

	if err := writeClaudeSettingsRoot(dir, path, root); err != nil {
		return "", err
	}

	return path, nil
}

func readClaudeSettingsRoot(path string) map[string]json.RawMessage {
	root := map[string]json.RawMessage{}

	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &root)
	}

	return root
}

func writeClaudeSettingsRoot(dir, path string, root map[string]json.RawMessage) error {
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
