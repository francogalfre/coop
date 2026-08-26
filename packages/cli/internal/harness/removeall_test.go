package harness_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/harness"
	"github.com/francogalfre/coop/packages/cli/internal/harness/claudecode"
	"github.com/francogalfre/coop/packages/cli/internal/harness/generic"
	"github.com/francogalfre/coop/packages/cli/internal/harness/opencode"
	"github.com/francogalfre/coop/packages/cli/internal/harness/pi"
)

func TestRemoveAllTracesRemovesEveryAdaptersTraces(t *testing.T) {
	dir := t.TempDir()

	if _, err := (claudecode.Adapter{}).Install(dir, "http://127.0.0.1:12345"); err != nil {
		t.Fatalf("claudecode Install() error = %v", err)
	}
	if _, err := (opencode.Adapter{}).Install(dir, "http://127.0.0.1:12345"); err != nil {
		t.Fatalf("opencode Install() error = %v", err)
	}
	if _, err := (pi.Adapter{}).Install(dir, "http://127.0.0.1:12345"); err != nil {
		t.Fatalf("pi Install() error = %v", err)
	}

	adapters := []harness.Adapter{claudecode.Adapter{}, opencode.Adapter{}, pi.Adapter{}, generic.Adapter{}}

	if err := harness.RemoveAllTraces(dir, adapters); err != nil {
		t.Fatalf("RemoveAllTraces() error = %v", err)
	}

	for _, p := range []string{
		filepath.Join(dir, ".claude", "settings.local.json"),
		filepath.Join(dir, ".opencode", "plugin", "coop.js"),
		filepath.Join(dir, ".pi", "extensions", "coop.ts"),
	} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s still exists after RemoveAllTraces(), err = %v", p, err)
		}
	}
}

func TestRemoveAllTracesNoopWhenNothingInstalled(t *testing.T) {
	dir := t.TempDir()

	adapters := []harness.Adapter{claudecode.Adapter{}, opencode.Adapter{}, pi.Adapter{}, generic.Adapter{}}

	if err := harness.RemoveAllTraces(dir, adapters); err != nil {
		t.Fatalf("RemoveAllTraces() error = %v, want nil for a directory with nothing installed", err)
	}
}
