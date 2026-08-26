package main

import (
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

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

func TestRequireLoginForProjectFailsWithoutCredential(t *testing.T) {
	err := requireLoginForProject("attach", config.Config{}, "my-project")
	if err == nil {
		t.Fatal("expected error when --project is set without a CLI credential")
	}
}

func TestRequireLoginForProjectPassesWithCredential(t *testing.T) {
	err := requireLoginForProject("attach", config.Config{CLICredential: "abc123"}, "my-project")
	if err != nil {
		t.Fatalf("requireLoginForProject: %v", err)
	}
}

func TestRequireLoginForProjectPassesWithoutProject(t *testing.T) {
	err := requireLoginForProject("attach", config.Config{}, "")
	if err != nil {
		t.Fatalf("requireLoginForProject: %v", err)
	}
}
