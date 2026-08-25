package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpencodeAdapterInstallWritesPluginWithBaseURL(t *testing.T) {
	dir := t.TempDir()

	inst, err := opencodeAdapter{}.Install(dir, "http://127.0.0.1:8788")
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	wantPath := filepath.Join(dir, ".opencode", "plugin", "coop.js")
	if len(inst.Paths) != 1 || inst.Paths[0] != wantPath {
		t.Fatalf("got Paths %v, want [%s]", inst.Paths, wantPath)
	}

	contents, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read installed plugin: %v", err)
	}

	if strings.Contains(string(contents), "{{BASE_URL}}") {
		t.Error("plugin still contains the template placeholder")
	}
	if !strings.Contains(string(contents), "http://127.0.0.1:8788") {
		t.Error("plugin does not contain the base URL")
	}

	if err := inst.Remove(); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if _, err := os.Stat(wantPath); !os.IsNotExist(err) {
		t.Fatalf("plugin file still exists after Remove(), err = %v", err)
	}
}
