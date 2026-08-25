package harness

import (
	"errors"
	"path/filepath"
)

// RemoveAll strips every trace coop may have left in dir for any harness,
// regardless of whether that harness is currently on PATH -- this backs
// `coop detach`, the disaster-recovery command for a killed `coop attach`,
// where the environment that ran the crashed process may differ from the
// one running detach.
func RemoveAll(dir string) error {
	return errors.Join(
		removeClaudeSettings(dir),
		removeIfExists(filepath.Join(dir, ".opencode", "plugin", "coop.js")),
		removeIfExists(filepath.Join(dir, ".pi", "extensions", "coop.ts")),
	)
}
