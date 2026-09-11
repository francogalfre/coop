package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/dirregistry"
	"github.com/francogalfre/coop/packages/cli/internal/harness"
	"github.com/francogalfre/coop/packages/cli/internal/harness/claudecode"
	"github.com/francogalfre/coop/packages/cli/internal/harness/opencode"
	"github.com/francogalfre/coop/packages/cli/internal/harness/pi"
)

func TestRejectConflictingAttachFailsForSecondOpencode(t *testing.T) {
	dir := t.TempDir()

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-a",
		PID:       os.Getpid(),
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{opencode.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err := rejectConflictingAttach(dir, []harness.Adapter{opencode.Adapter{}})
	if err == nil {
		t.Fatal("second opencode attach was allowed, want explicit error")
	}

	for _, want := range []string{"opencode", "sess-a", "one coop session per directory"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestRejectConflictingAttachFailsForSecondPi(t *testing.T) {
	dir := t.TempDir()

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-a",
		PID:       os.Getpid(),
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{pi.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := rejectConflictingAttach(dir, []harness.Adapter{pi.Adapter{}}); err == nil {
		t.Fatal("second pi attach was allowed, want explicit error")
	}
}

func TestRejectConflictingAttachAllowsSecondClaudeCode(t *testing.T) {
	dir := t.TempDir()

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-a",
		PID:       os.Getpid(),
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{claudecode.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := rejectConflictingAttach(dir, []harness.Adapter{claudecode.Adapter{}}); err != nil {
		t.Fatalf("second claude-code attach rejected: %v", err)
	}
}

func TestRejectConflictingAttachAllowsDifferentSingleFileHarness(t *testing.T) {
	dir := t.TempDir()

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-a",
		PID:       os.Getpid(),
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{opencode.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := rejectConflictingAttach(dir, []harness.Adapter{pi.Adapter{}}); err != nil {
		t.Fatalf("pi attach rejected while opencode holds the directory: %v", err)
	}
}

func TestRejectConflictingAttachIgnoresDeadSession(t *testing.T) {
	dir := t.TempDir()

	cmd := exec.Command("true")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-dead",
		PID:       cmd.Process.Pid,
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{opencode.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := rejectConflictingAttach(dir, []harness.Adapter{opencode.Adapter{}}); err != nil {
		t.Fatalf("attach rejected by a dead session's stale entry: %v", err)
	}
}

func TestRejectConflictingAttachPassesOnEmptyDirectory(t *testing.T) {
	if err := rejectConflictingAttach(t.TempDir(), []harness.Adapter{opencode.Adapter{}}); err != nil {
		t.Fatalf("attach rejected in a fresh directory: %v", err)
	}
}

func TestRejectConflictingAttachFailsOnMalformedRegistry(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(dir+"/.coop", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/.coop/sessions.json", []byte(`{not json`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := rejectConflictingAttach(dir, nil); err == nil {
		t.Fatal("malformed sessions.json was swallowed, want explicit error")
	}
}
