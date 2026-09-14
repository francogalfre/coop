package dirregistry

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func liveEntry(sessionID string) Entry {
	return Entry{SessionID: sessionID, PID: os.Getpid(), BaseURL: "http://127.0.0.1:8788", Harnesses: []string{"claude-code"}}
}

func deadPID(t *testing.T) int {
	t.Helper()

	cmd := exec.Command("true")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start throwaway process: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait throwaway process: %v", err)
	}

	return cmd.Process.Pid
}

func TestLoadEmptyDirReturnsNoEntries(t *testing.T) {
	dir := t.TempDir()

	entries, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("got %d entries, want none", len(entries))
	}
}

func TestRegisterThenLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()

	first := liveEntry("sess-a")
	first.BaseURL = "http://127.0.0.1:8788"
	second := liveEntry("sess-b")
	second.BaseURL = "http://127.0.0.1:8789"
	second.Harnesses = []string{"opencode"}

	if err := Register(dir, first); err != nil {
		t.Fatalf("Register(first) error = %v", err)
	}
	if err := Register(dir, second); err != nil {
		t.Fatalf("Register(second) error = %v", err)
	}

	entries, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].SessionID != "sess-a" || entries[1].SessionID != "sess-b" {
		t.Fatalf("got sessions %v, want [sess-a sess-b]", entries)
	}
	if entries[1].Harnesses[0] != "opencode" {
		t.Fatalf("got harnesses %v, want [opencode]", entries[1].Harnesses)
	}
}

func TestRegisterReplacesSameSessionID(t *testing.T) {
	dir := t.TempDir()

	stale := liveEntry("sess-a")
	stale.BaseURL = "http://127.0.0.1:8788"

	if err := Register(dir, stale); err != nil {
		t.Fatalf("Register(stale) error = %v", err)
	}

	fresh := liveEntry("sess-a")
	fresh.BaseURL = "http://127.0.0.1:8789"

	if err := Register(dir, fresh); err != nil {
		t.Fatalf("Register(fresh) error = %v", err)
	}

	entries, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1 (same session re-registered)", len(entries))
	}
	if entries[0].BaseURL != "http://127.0.0.1:8789" {
		t.Fatalf("got base URL %q, want the fresh one", entries[0].BaseURL)
	}
}

func TestLoadPrunesDeadProcessEntries(t *testing.T) {
	dir := t.TempDir()

	if err := Register(dir, Entry{SessionID: "sess-dead", PID: deadPID(t), BaseURL: "http://127.0.0.1:8788"}); err != nil {
		t.Fatalf("Register(dead) error = %v", err)
	}
	if err := Register(dir, liveEntry("sess-live")); err != nil {
		t.Fatalf("Register(live) error = %v", err)
	}

	entries, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(entries) != 1 || entries[0].SessionID != "sess-live" {
		t.Fatalf("got entries %+v, want only sess-live", entries)
	}

	entries, err = Load(dir)
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("pruned entry came back: %+v", entries)
	}
}

func TestUnregisterRemovesOnlyThatSession(t *testing.T) {
	dir := t.TempDir()

	if err := Register(dir, liveEntry("sess-a")); err != nil {
		t.Fatalf("Register(sess-a) error = %v", err)
	}
	if err := Register(dir, liveEntry("sess-b")); err != nil {
		t.Fatalf("Register(sess-b) error = %v", err)
	}

	if err := Unregister(dir, "sess-a"); err != nil {
		t.Fatalf("Unregister(sess-a) error = %v", err)
	}

	entries, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(entries) != 1 || entries[0].SessionID != "sess-b" {
		t.Fatalf("got entries %+v, want only sess-b", entries)
	}
}

func TestUnregisterAbsentSessionIsNoop(t *testing.T) {
	dir := t.TempDir()

	if err := Register(dir, liveEntry("sess-a")); err != nil {
		t.Fatalf("Register(sess-a) error = %v", err)
	}

	if err := Unregister(dir, "sess-gone"); err != nil {
		t.Fatalf("Unregister(absent) error = %v, want nil", err)
	}

	entries, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("got entries %+v, want sess-a untouched", entries)
	}
}

func TestLoadMalformedFileFailsExplicitly(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, coopDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pathFor(dir), []byte(`{not json`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(dir); err == nil {
		t.Fatal("Load() on malformed sessions.json succeeded, want explicit error")
	}
}

func TestClaimRejectsConflictAtomically(t *testing.T) {
	dir := t.TempDir()

	if err := Register(dir, liveEntry("sess-a")); err != nil {
		t.Fatalf("Register(sess-a) error = %v", err)
	}

	conflictErr := func(existing []Entry) error {
		if len(existing) > 0 {
			return os.ErrExist
		}
		return nil
	}

	if err := Claim(dir, liveEntry("sess-b"), conflictErr, noInstall); err == nil {
		t.Fatal("Claim succeeded, want conflict error")
	}

	entries, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1 (sess-b should not have been registered)", len(entries))
	}
}

func TestClaimAllowsNoConflict(t *testing.T) {
	dir := t.TempDir()

	noConflict := func(existing []Entry) error { return nil }

	if err := Claim(dir, liveEntry("sess-a"), noConflict, noInstall); err != nil {
		t.Fatalf("Claim error = %v", err)
	}

	entries, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
}

func noInstall() error { return nil }
