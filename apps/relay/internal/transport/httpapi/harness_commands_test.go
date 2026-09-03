package httpapi

import (
	"os"
	"regexp"
	"testing"
)

func TestHarnessCommandsMatchesTheProtocolAllowlist(t *testing.T) {
	data, err := os.ReadFile("../../../../../packages/protocol/src/shared/commands.ts")
	if err != nil {
		t.Skipf("could not read protocol source to cross-check the allowlist: %v", err)
	}

	match := regexp.MustCompile(`HARNESS_COMMANDS = \[([^\]]+)\]`).FindSubmatch(data)
	if match == nil {
		t.Fatal("could not find HARNESS_COMMANDS in packages/protocol/src/shared/commands.ts")
	}

	names := regexp.MustCompile(`"([a-z]+)"`).FindAllSubmatch(match[1], -1)
	if len(names) != len(harnessCommands) {
		t.Fatalf("got %d commands in the Go allowlist, want %d to match the protocol", len(harnessCommands), len(names))
	}

	for _, m := range names {
		name := string(m[1])
		if !harnessCommands[name] {
			t.Errorf("protocol declares %q, missing from the relay's allowlist", name)
		}
	}
}
