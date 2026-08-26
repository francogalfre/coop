package pi

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/francogalfre/coop/packages/cli/internal/harness"
)

const Name = "pi"

//go:embed assets/pi.ts
var extensionTemplate string

type Adapter struct{}

func (Adapter) Name() string { return Name }

func (Adapter) Binary() string { return "pi" }

func (Adapter) IsFallback() bool { return false }

func (Adapter) Detect(dir string) bool {
	_, err := exec.LookPath("pi")
	return err == nil
}

func (Adapter) Install(dir, baseURL string) (harness.Installation, error) {
	path := extensionPath(dir)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return harness.Installation{}, fmt.Errorf("harness: create %s: %w", filepath.Dir(path), err)
	}

	contents := strings.ReplaceAll(extensionTemplate, "{{BASE_URL}}", baseURL)

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		return harness.Installation{}, fmt.Errorf("harness: write %s: %w", path, err)
	}

	return harness.Installation{
		Paths:  []string{path},
		Remove: func() error { return removeIfExists(path) },
	}, nil
}

func (Adapter) RemoveAll(dir string) error {
	return removeIfExists(extensionPath(dir))
}

func extensionPath(dir string) string {
	return filepath.Join(dir, ".pi", "extensions", "coop.ts")
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("harness: remove %s: %w", path, err)
	}

	return nil
}
