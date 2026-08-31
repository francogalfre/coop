package main

import (
	"io"
	"os"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

func TestRunVersion(t *testing.T) {
	for _, arg := range []string{"version", "--version", "-v"} {
		out := captureStdout(t, func() {
			if err := run([]string{arg}); err != nil {
				t.Fatalf("run(%q): %v", arg, err)
			}
		})

		want := "coop " + version + "\n"
		if out != want {
			t.Fatalf("run(%q) printed %q, want %q", arg, out, want)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	orig := os.Stdout
	os.Stdout = w
	fn()
	os.Stdout = orig
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}

	return string(out)
}

func TestParseAttachRunFlagsProject(t *testing.T) {
	_, project, remaining, err := parseAttachRunFlags("attach", []string{"--project=my-project"})
	if err != nil {
		t.Fatalf("parseAttachRunFlags: %v", err)
	}

	if project != "my-project" {
		t.Fatalf("got project %q, want %q", project, "my-project")
	}

	if len(remaining) != 0 {
		t.Fatalf("got remaining %v, want none", remaining)
	}
}

func TestParseAttachRunFlagsHarnessAndProject(t *testing.T) {
	harnessFlag, project, remaining, err := parseAttachRunFlags("attach", []string{"--harness=claude-code", "--project=my-project"})
	if err != nil {
		t.Fatalf("parseAttachRunFlags: %v", err)
	}

	if harnessFlag != "claude-code" {
		t.Fatalf("got harness %q, want %q", harnessFlag, "claude-code")
	}

	if project != "my-project" {
		t.Fatalf("got project %q, want %q", project, "my-project")
	}

	if len(remaining) != 0 {
		t.Fatalf("got remaining %v, want none", remaining)
	}
}

func TestParseAttachRunFlagsAbsent(t *testing.T) {
	_, project, remaining, err := parseAttachRunFlags("run", []string{"--", "claude", "--flag"})
	if err != nil {
		t.Fatalf("parseAttachRunFlags: %v", err)
	}

	if project != "" {
		t.Fatalf("got project %q, want empty", project)
	}

	if len(remaining) != 2 || remaining[0] != "claude" || remaining[1] != "--flag" {
		t.Fatalf("got remaining %v, want [claude --flag]", remaining)
	}
}

func TestParseAttachRunFlagsWithRunArgs(t *testing.T) {
	_, project, remaining, err := parseAttachRunFlags("run", []string{"--project=my-project", "--", "claude"})
	if err != nil {
		t.Fatalf("parseAttachRunFlags: %v", err)
	}

	if project != "my-project" {
		t.Fatalf("got project %q, want %q", project, "my-project")
	}

	if len(remaining) != 1 || remaining[0] != "claude" {
		t.Fatalf("got remaining %v, want [claude]", remaining)
	}
}

func TestRequireProjectAndLoginFailsWithoutCredential(t *testing.T) {
	err := requireProjectAndLogin("attach", config.Config{}, "my-project")
	if err == nil {
		t.Fatal("expected error when --project is set without a CLI credential")
	}
}

func TestRequireProjectAndLoginPassesWithCredential(t *testing.T) {
	err := requireProjectAndLogin("attach", config.Config{CLICredential: "abc123"}, "my-project")
	if err != nil {
		t.Fatalf("requireProjectAndLogin: %v", err)
	}
}

func TestRequireProjectAndLoginFailsWithoutProject(t *testing.T) {
	err := requireProjectAndLogin("attach", config.Config{CLICredential: "abc123"}, "")
	if err == nil {
		t.Fatal("expected error when --project is missing")
	}
}
