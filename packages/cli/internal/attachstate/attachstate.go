package attachstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Record is what `coop attach` leaves behind so `coop detach` can end the
// session even when attach was killed without running its own cleanup.
type Record struct {
	SessionID string `json:"session_id"`
	RelayURL  string `json:"relay_url"`
	Project   string `json:"project"`
}

func dirFor() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("attachstate: user home dir: %w", err)
	}

	return filepath.Join(home, ".config", "coop", "attached"), nil
}

func pathFor(workdir string) (string, error) {
	abs, err := filepath.Abs(workdir)
	if err != nil {
		return "", fmt.Errorf("attachstate: resolve %q: %w", workdir, err)
	}

	dir, err := dirFor()
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte(abs))

	return filepath.Join(dir, hex.EncodeToString(sum[:])+".json"), nil
}

func Save(workdir string, rec Record) error {
	path, err := pathFor(workdir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("attachstate: create dir: %w", err)
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("attachstate: marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("attachstate: write %s: %w", path, err)
	}

	return nil
}

func Load(workdir string) (Record, bool, error) {
	path, err := pathFor(workdir)
	if err != nil {
		return Record{}, false, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Record{}, false, nil
		}

		return Record{}, false, fmt.Errorf("attachstate: read %s: %w", path, err)
	}

	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return Record{}, false, fmt.Errorf("attachstate: parse %s: %w", path, err)
	}

	return rec, true, nil
}

func Remove(workdir string) error {
	path, err := pathFor(workdir)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("attachstate: remove %s: %w", path, err)
	}

	return nil
}
