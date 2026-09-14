package dirregistry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Entry is one active `coop attach` recorded in a working directory, so a
// later attach can detect it and shutdown can remove exactly its own record.
type Entry struct {
	SessionID string   `json:"session_id"`
	PID       int      `json:"pid"`
	BaseURL   string   `json:"base_url"`
	Harnesses []string `json:"harnesses"`
}

type file struct {
	Sessions []Entry `json:"sessions"`
}

const (
	coopDirName      = ".coop"
	sessionsFileName = "sessions.json"
	sessionsLockName = "sessions.lock"
)

func pathFor(dir string) string {
	return filepath.Join(dir, coopDirName, sessionsFileName)
}

// Load returns the active entries and prunes records whose process is gone,
// so a crashed attach stops blocking new ones.
func Load(dir string) ([]Entry, error) {
	if _, err := os.Stat(pathFor(dir)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("dirregistry: stat %s: %w", pathFor(dir), err)
	}

	var entries []Entry

	err := locked(dir, func() error {
		live, changed, err := readLive(dir)
		if err != nil {
			return err
		}

		entries = live

		if changed {
			return persist(dir, live)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func Register(dir string, e Entry) error {
	return locked(dir, func() error {
		live, _, err := readLive(dir)
		if err != nil {
			return err
		}

		return registerLocked(dir, live, e)
	})
}

// Claim runs conflict, install, and the registry write under one directory
// lock. Harnesses like opencode share a single plugin file, so an install that
// raced a conflicting attach would overwrite that attach's file and then delete
// it while rolling back.
func Claim(dir string, e Entry, conflict func(existing []Entry) error, install func() error) error {
	return locked(dir, func() error {
		live, _, err := readLive(dir)
		if err != nil {
			return err
		}

		// A crash-restart reuses its session id, so it must not conflict with itself.
		others := make([]Entry, 0, len(live))
		for _, existing := range live {
			if existing.SessionID != e.SessionID {
				others = append(others, existing)
			}
		}

		if err := conflict(others); err != nil {
			return err
		}

		if err := install(); err != nil {
			return err
		}

		return registerLocked(dir, live, e)
	})
}

func registerLocked(dir string, live []Entry, e Entry) error {
	kept := make([]Entry, 0, len(live)+1)
	for _, existing := range live {
		if existing.SessionID != e.SessionID {
			kept = append(kept, existing)
		}
	}

	return persist(dir, append(kept, e))
}

func Unregister(dir, sessionID string) error {
	if _, err := os.Stat(pathFor(dir)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("dirregistry: stat %s: %w", pathFor(dir), err)
	}

	return locked(dir, func() error {
		live, _, err := readLive(dir)
		if err != nil {
			return err
		}

		kept := make([]Entry, 0, len(live))
		for _, e := range live {
			if e.SessionID != sessionID {
				kept = append(kept, e)
			}
		}

		return persist(dir, kept)
	})
}

func readLive(dir string) ([]Entry, bool, error) {
	path := pathFor(dir)

	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("dirregistry: read %s: %w", path, err)
	}

	var f file
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, false, fmt.Errorf("dirregistry: parse %s: %w", path, err)
	}

	live := make([]Entry, 0, len(f.Sessions))
	for _, e := range f.Sessions {
		if processAlive(e.PID) {
			live = append(live, e)
		}
	}

	return live, len(live) != len(f.Sessions), nil
}

func persist(dir string, entries []Entry) error {
	coopDir := filepath.Join(dir, coopDirName)

	if err := os.MkdirAll(coopDir, 0o755); err != nil {
		return fmt.Errorf("dirregistry: create %s: %w", coopDir, err)
	}

	out, err := json.MarshalIndent(file{Sessions: entries}, "", "  ")
	if err != nil {
		return fmt.Errorf("dirregistry: marshal: %w", err)
	}

	path := pathFor(dir)
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		return fmt.Errorf("dirregistry: write %s: %w", path, err)
	}

	return nil
}

// locked serializes registry mutations with flock, which the kernel releases
// even when an attach is killed mid-write, so no stale lock can wedge the
// directory.
func locked(dir string, fn func() error) error {
	coopDir := filepath.Join(dir, coopDirName)

	if err := os.MkdirAll(coopDir, 0o755); err != nil {
		return fmt.Errorf("dirregistry: create %s: %w", coopDir, err)
	}

	f, err := os.OpenFile(filepath.Join(coopDir, sessionsLockName), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("dirregistry: open lock: %w", err)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("dirregistry: lock %s: %w", f.Name(), err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	return fn()
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	return proc.Signal(syscall.Signal(0)) == nil
}
