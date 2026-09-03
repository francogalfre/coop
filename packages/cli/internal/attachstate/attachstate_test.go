package attachstate

import (
	"os"
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
