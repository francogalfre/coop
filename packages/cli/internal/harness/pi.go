package harness

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const piName = "pi"

//go:embed assets/pi.ts
var piExtensionTemplate string

type piAdapter struct{}

func (piAdapter) Name() string { return piName }

func (piAdapter) Detect(dir string) bool {
	_, err := exec.LookPath("pi")
	return err == nil
}

// Install writes into .pi/extensions/, which pi auto-loads via jiti once
// the user trusts the project interactively in pi's TUI -- there is no
// non-interactive way to grant that trust from here.
func (piAdapter) Install(dir, baseURL string) (Installation, error) {
	extDir := filepath.Join(dir, ".pi", "extensions")
	path := filepath.Join(extDir, "coop.ts")

	if err := os.MkdirAll(extDir, 0o755); err != nil {
		return Installation{}, fmt.Errorf("harness: create %s: %w", extDir, err)
	}

	contents := strings.ReplaceAll(piExtensionTemplate, "{{BASE_URL}}", baseURL)

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		return Installation{}, fmt.Errorf("harness: write %s: %w", path, err)
	}

	return Installation{
		Paths:  []string{path},
		Remove: func() error { return removeIfExists(path) },
	}, nil
}
