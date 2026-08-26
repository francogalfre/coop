package repoid

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDetectFallsBackToDirBaseNameOutsideGit(t *testing.T) {
	dir := t.TempDir()

	got := Detect(dir)
	want := filepath.Base(dir)

	if got != want {
		t.Fatalf("Detect(%q) = %q, want %q", dir, got, want)
	}
}

func TestDetectFallsBackToToplevelWhenNoRemote(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")

	got := Detect(dir)
	want := filepath.Base(dir)

	if got != want {
		t.Fatalf("Detect(%q) = %q, want %q", dir, got, want)
	}
}

func TestDetectNormalizesSSHRemote(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "remote", "add", "origin", "git@github.com:francogalfre/coop.git")

	got := Detect(dir)
	want := "github.com/francogalfre/coop"

	if got != want {
		t.Fatalf("Detect(%q) = %q, want %q", dir, got, want)
	}
}

func TestDetectNormalizesHTTPSRemote(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "remote", "add", "origin", "https://github.com/francogalfre/coop.git")

	got := Detect(dir)
	want := "github.com/francogalfre/coop"

	if got != want {
		t.Fatalf("Detect(%q) = %q, want %q", dir, got, want)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
