package attachstate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRemoveRoundTrips(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()

	want := Record{SessionID: "sess-abc", RelayURL: "http://localhost:8787", Project: "demo"}

	if err := Save(dir, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok, err := Load(dir)
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}

	if err := Remove(dir); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if _, ok, err := Load(dir); ok || err != nil {
		t.Fatalf("Load after Remove: ok=%v err=%v", ok, err)
	}
}

func TestLoadMissingIsNotAnError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, ok, err := Load(t.TempDir())
	if ok || err != nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestRemoveMissingIsNotAnError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := Remove(t.TempDir()); err != nil {
		t.Fatalf("Remove: %v", err)
	}
}

func TestSeparateDirsDoNotCollide(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a, b := t.TempDir(), t.TempDir()

	if err := Save(a, Record{SessionID: "sess-a"}); err != nil {
		t.Fatalf("Save a: %v", err)
	}
	if err := Save(b, Record{SessionID: "sess-b"}); err != nil {
		t.Fatalf("Save b: %v", err)
	}

	got, _, _ := Load(a)
	if got.SessionID != "sess-a" {
		t.Fatalf("dir a got %q", got.SessionID)
	}
}

func TestSaveWritesRestrictivePermissions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()

	if err := Save(dir, Record{SessionID: "sess-x"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path, err := pathFor(dir)
	if err != nil {
		t.Fatalf("pathFor: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestSaveAppendsMultipleSessions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()

	a := Record{SessionID: "sess-a", RelayURL: "http://a", Project: "p"}
	b := Record{SessionID: "sess-b", RelayURL: "http://b", Project: "p"}

	if err := Save(dir, a); err != nil {
		t.Fatalf("Save a: %v", err)
	}
	if err := Save(dir, b); err != nil {
		t.Fatalf("Save b: %v", err)
	}

	all, ok, err := LoadAll(dir)
	if err != nil || !ok {
		t.Fatalf("LoadAll: ok=%v err=%v", ok, err)
	}
	if len(all) != 2 {
		t.Fatalf("got %d records, want 2", len(all))
	}

	latest, _, _ := Load(dir)
	if latest.SessionID != "sess-b" {
		t.Fatalf("Load latest = %q, want sess-b", latest.SessionID)
	}
}

func TestSaveReplacesDuplicateSession(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()

	if err := Save(dir, Record{SessionID: "sess-x", RelayURL: "http://old"}); err != nil {
		t.Fatalf("Save old: %v", err)
	}
	if err := Save(dir, Record{SessionID: "sess-x", RelayURL: "http://new"}); err != nil {
		t.Fatalf("Save new: %v", err)
	}

	all, _, _ := LoadAll(dir)
	if len(all) != 1 {
		t.Fatalf("got %d records, want 1", len(all))
	}
	if all[0].RelayURL != "http://new" {
		t.Fatalf("RelayURL = %q, want http://new", all[0].RelayURL)
	}
}

func TestRemoveSessionKeepsOtherSessions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()

	if err := Save(dir, Record{SessionID: "sess-a"}); err != nil {
		t.Fatalf("Save a: %v", err)
	}
	if err := Save(dir, Record{SessionID: "sess-b"}); err != nil {
		t.Fatalf("Save b: %v", err)
	}

	if err := RemoveSession(dir, "sess-a"); err != nil {
		t.Fatalf("RemoveSession: %v", err)
	}

	all, _, _ := LoadAll(dir)
	if len(all) != 1 || all[0].SessionID != "sess-b" {
		t.Fatalf("after remove got %+v, want [sess-b]", all)
	}
}

func TestRemoveSessionRemovesFileWhenEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()

	if err := Save(dir, Record{SessionID: "sess-x"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := RemoveSession(dir, "sess-x"); err != nil {
		t.Fatalf("RemoveSession: %v", err)
	}

	path, _ := pathFor(dir)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("state file still exists after removing only session")
	}
}

func TestLoadMigratesLegacySingleRecord(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()

	path, err := pathFor(dir)
	if err != nil {
		t.Fatalf("pathFor: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	legacy := `{"session_id":"legacy-sess","relay_url":"http://old","project":"demo"}`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	rec, ok, err := Load(dir)
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}
	if rec.SessionID != "legacy-sess" {
		t.Fatalf("got %q, want legacy-sess", rec.SessionID)
	}

	all, _, _ := LoadAll(dir)
	if len(all) != 1 {
		t.Fatalf("LoadAll after legacy migration = %d records, want 1", len(all))
	}
}
