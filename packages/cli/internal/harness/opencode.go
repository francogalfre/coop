package harness

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const opencodeName = "opencode"

//go:embed assets/opencode.js
var opencodePluginTemplate string

type opencodeAdapter struct{}

func (opencodeAdapter) Name() string { return opencodeName }

func (opencodeAdapter) Detect(dir string) bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}

// Install writes into .opencode/plugin/, which opencode auto-loads without
// needing an opencode.json entry -- no JSON surgery on the project's config.
func (opencodeAdapter) Install(dir, baseURL string) (Installation, error) {
	pluginDir := filepath.Join(dir, ".opencode", "plugin")
	path := filepath.Join(pluginDir, "coop.js")

	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return Installation{}, fmt.Errorf("harness: create %s: %w", pluginDir, err)
	}

	contents := strings.ReplaceAll(opencodePluginTemplate, "{{BASE_URL}}", baseURL)

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		return Installation{}, fmt.Errorf("harness: write %s: %w", path, err)
	}

	return Installation{
		Paths:  []string{path},
		Remove: func() error { return removeIfExists(path) },
	}, nil
}
