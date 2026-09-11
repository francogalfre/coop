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

// Save appends a record to the directory's state file, replacing any existing
// record with the same session id so a restarted attach on the same session
// does not accumulate duplicates.
func Save(workdir string, rec Record) error {
	path, err := pathFor(workdir)
	if err != nil {
		return err
	}

	records, _, err := loadAll(workdir)
	if err != nil {
		return err
	}

	records = appendWithoutDuplicate(records, rec)

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("attachstate: create dir: %w", err)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("attachstate: marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("attachstate: write %s: %w", path, err)
	}

	return nil
}

func appendWithoutDuplicate(records []Record, rec Record) []Record {
	kept := make([]Record, 0, len(records)+1)
	for _, existing := range records {
		if existing.SessionID != rec.SessionID {
			kept = append(kept, existing)
		}
	}

	return append(kept, rec)
}

// Load returns the most recent record for the working directory.
// It exists for callers that expect a single attach per directory.
func Load(workdir string) (Record, bool, error) {
	records, ok, err := loadAll(workdir)
	if err != nil || !ok || len(records) == 0 {
		return Record{}, false, err
	}

	return records[len(records)-1], true, nil
}

// LoadAll returns every attach record stored for the working directory.
func LoadAll(workdir string) ([]Record, bool, error) {
	return loadAll(workdir)
}

func loadAll(workdir string) ([]Record, bool, error) {
	path, err := pathFor(workdir)
	if err != nil {
		return nil, false, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("attachstate: read %s: %w", path, err)
	}

	// Previous versions of coop stored a single Record object. Migrate on read
	// so an in-place upgrade does not leave stale state behind.
	if records, ok := tryLoadLegacy(data); ok {
		return []Record{records}, true, nil
	}

	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, false, fmt.Errorf("attachstate: parse %s: %w", path, err)
	}

	return records, true, nil
}

func tryLoadLegacy(data []byte) (Record, bool) {
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return Record{}, false
	}

	// A legacy single record has no array brackets. A one-element array would
	// also unmarshal into the struct, so check the first non-space byte.
	for _, b := range data {
		if b == ' ' || b == '\n' || b == '\r' || b == '\t' {
			continue
		}
		return rec, b == '{'
	}

	return Record{}, false
}

// RemoveSession deletes one session's record from the directory state file.
// If no records remain, the file is removed.
func RemoveSession(workdir, sessionID string) error {
	path, err := pathFor(workdir)
	if err != nil {
		return err
	}

	records, ok, err := loadAll(workdir)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	kept := make([]Record, 0, len(records))
	for _, rec := range records {
		if rec.SessionID != sessionID {
			kept = append(kept, rec)
		}
	}

	if len(kept) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("attachstate: remove %s: %w", path, err)
		}
		return nil
	}

	data, err := json.MarshalIndent(kept, "", "  ")
	if err != nil {
		return fmt.Errorf("attachstate: marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("attachstate: write %s: %w", path, err)
	}

	return nil
}

// Remove deletes the entire state file for the working directory.
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
