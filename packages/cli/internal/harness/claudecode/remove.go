package claudecode

import (
	"encoding/json"
	"fmt"
	"os"
)

func removeSettings(dir string) error {
	return removeCoopEntries(dir, func(event string, e hookEntry) bool {
		return isCoopEntry(e)
	})
}

func removeSettingsFor(dir, baseURL string) error {
	return removeCoopEntries(dir, func(event string, e hookEntry) bool {
		return isCoopEntry(e) && e.URL == coopURL(baseURL, event)
	})
}

func removeCoopEntries(dir string, drop func(event string, e hookEntry) bool) error {
	path := settingsPath(dir)

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
		remaining := keepGroups(groups, event, drop)
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

	return writeSettingsRoot(dir, path, root)
}

func keepGroups(groups []hookGroup, event string, drop func(event string, e hookEntry) bool) []hookGroup {
	kept := make([]hookGroup, 0, len(groups))

	for _, g := range groups {
		entries := make([]hookEntry, 0, len(g.Hooks))

		for _, e := range g.Hooks {
			if !drop(event, e) {
				entries = append(entries, e)
			}
		}

		if len(entries) > 0 {
			kept = append(kept, hookGroup{Hooks: entries, extra: g.extra})
		}
	}

	return kept
}
