package main

import (
	"errors"
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

func TestClaimDirectoryFailsForSecondOpencode(t *testing.T) {
	dir := t.TempDir()

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-a",
		PID:       os.Getpid(),
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{opencode.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	err := claimFor(dir, []harness.Adapter{opencode.Adapter{}})
	if err == nil {
		t.Fatal("second opencode attach was allowed, want explicit error")
	}

	for _, want := range []string{"opencode", "sess-a", "one coop session per directory"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestClaimDirectoryFailsForSecondPi(t *testing.T) {
	dir := t.TempDir()

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-a",
		PID:       os.Getpid(),
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{pi.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := claimFor(dir, []harness.Adapter{pi.Adapter{}}); err == nil {
		t.Fatal("second pi attach was allowed, want explicit error")
	}
}

func TestClaimDirectoryAllowsSecondClaudeCode(t *testing.T) {
	dir := t.TempDir()

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-a",
		PID:       os.Getpid(),
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{claudecode.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := claimFor(dir, []harness.Adapter{claudecode.Adapter{}}); err != nil {
		t.Fatalf("second claude-code attach rejected: %v", err)
	}
}

func TestClaimDirectoryAllowsDifferentSingleFileHarness(t *testing.T) {
	dir := t.TempDir()

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-a",
		PID:       os.Getpid(),
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{opencode.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := claimFor(dir, []harness.Adapter{pi.Adapter{}}); err != nil {
		t.Fatalf("pi attach rejected while opencode holds the directory: %v", err)
	}
}

func TestClaimDirectoryIgnoresDeadSession(t *testing.T) {
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

	if err := claimFor(dir, []harness.Adapter{opencode.Adapter{}}); err != nil {
		t.Fatalf("attach rejected by a dead session's stale entry: %v", err)
	}
}

func TestClaimDirectoryPassesOnEmptyDirectory(t *testing.T) {
	if err := claimFor(t.TempDir(), []harness.Adapter{opencode.Adapter{}}); err != nil {
		t.Fatalf("attach rejected in a fresh directory: %v", err)
	}
}

func TestClaimDirectoryFailsOnMalformedRegistry(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(dir+"/.coop", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/.coop/sessions.json", []byte(`{not json`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := claimFor(dir, nil); err == nil {
		t.Fatal("malformed sessions.json was swallowed, want explicit error")
	}
}

func claimFor(dir string, adapters []harness.Adapter) error {
	return claimDirectory(dir, "sess-new", "http://127.0.0.1:9999", adapters, func() error { return nil })
}

func TestClaimDirectorySkipsInstallWhenAnotherSessionHoldsTheHarness(t *testing.T) {
	dir := t.TempDir()

	if err := dirregistry.Register(dir, dirregistry.Entry{
		SessionID: "sess-a",
		PID:       os.Getpid(),
		BaseURL:   "http://127.0.0.1:8788",
		Harnesses: []string{opencode.Name},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	installed := false
	err := claimDirectory(dir, "sess-b", "http://127.0.0.1:9999",
		[]harness.Adapter{opencode.Adapter{}}, func() error {
			installed = true
			return nil
		})

	if err == nil {
		t.Fatal("second opencode attach was allowed, want explicit error")
	}
	if installed {
		t.Fatal("install ran for a rejected attach, overwriting the holder's plugin file")
	}
}

func TestClaimDirectoryLeavesNoEntryWhenInstallFails(t *testing.T) {
	dir := t.TempDir()

	err := claimDirectory(dir, "sess-a", "http://127.0.0.1:9999",
		[]harness.Adapter{opencode.Adapter{}}, func() error {
			return errInstallFailed
		})
	if !errors.Is(err, errInstallFailed) {
		t.Fatalf("got %v, want the install error to surface", err)
	}

	active, err := dirregistry.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("got %d registered sessions after a failed install, want 0", len(active))
	}
}

var errInstallFailed = errors.New("install failed")
