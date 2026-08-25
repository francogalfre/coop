package harness

import (
	"encoding/json"
	"fmt"
	"os"
)

// removeClaudeSettings is the exact inverse of writeClaudeSettings: it
// strips every coop entry, drops emptied groups and event keys, drops
// "hooks" if empty, and deletes the settings file entirely if nothing
// is left.
func removeClaudeSettings(dir string) error {
	path := claudeSettingsPath(dir)

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("harness: read %s: %w", path, err)
	}

	root := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("harness: unmarshal %s: %w", path, err)
	}

	hooksRoot := map[string][]hookGroup{}
	if hooksRaw, ok := root["hooks"]; ok {
		if err := json.Unmarshal(hooksRaw, &hooksRoot); err != nil {
			return fmt.Errorf("harness: unmarshal hooks: %w", err)
		}
	}

	for event, groups := range hooksRoot {
		remaining := withoutCoopEntries(groups)
		if len(remaining) == 0 {
			delete(hooksRoot, event)
		} else {
			hooksRoot[event] = remaining
		}
	}

	if len(hooksRoot) == 0 {
		delete(root, "hooks")
	} else {
		hooksJSON, err := json.Marshal(hooksRoot)
		if err != nil {
			return fmt.Errorf("harness: marshal hooks: %w", err)
		}
		root["hooks"] = hooksJSON
	}

	if len(root) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("harness: remove %s: %w", path, err)
		}
		return nil
	}

	return writeClaudeSettingsRoot(dir, path, root)
}
