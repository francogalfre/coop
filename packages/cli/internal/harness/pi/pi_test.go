package pi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdapterInstallWritesExtensionWithBaseURL(t *testing.T) {
	dir := t.TempDir()

	inst, err := Adapter{}.Install(dir, "http://127.0.0.1:8788")
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	wantPath := filepath.Join(dir, ".pi", "extensions", "coop.ts")
	if len(inst.Paths) != 1 || inst.Paths[0] != wantPath {
		t.Fatalf("got Paths %v, want [%s]", inst.Paths, wantPath)
	}

	contents, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read installed extension: %v", err)
	}

	if strings.Contains(string(contents), "{{BASE_URL}}") {
		t.Error("extension still contains the template placeholder")
	}
	if !strings.Contains(string(contents), "http://127.0.0.1:8788") {
		t.Error("extension does not contain the base URL")
	}

	if err := inst.Remove(); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if _, err := os.Stat(wantPath); !os.IsNotExist(err) {
		t.Fatalf("extension file still exists after Remove(), err = %v", err)
	}
}

func TestAdapterRemoveAllRemovesExtensionWithoutInstall(t *testing.T) {
	dir := t.TempDir()

	if _, err := (Adapter{}).Install(dir, "http://127.0.0.1:8788"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if err := (Adapter{}).RemoveAll(dir); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}

	if _, err := os.Stat(extensionPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("extension file still exists after RemoveAll(), err = %v", err)
	}
}

func TestAdapterRemoveAllNoopWhenNothingInstalled(t *testing.T) {
	if err := (Adapter{}).RemoveAll(t.TempDir()); err != nil {
		t.Fatalf("RemoveAll() error = %v, want nil", err)
	}
}
