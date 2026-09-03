package commands

import "strings"

// Allowed mirrors packages/protocol/src/shared/commands.ts's HARNESS_COMMANDS.
// A harness command is typed into the session owner's own terminal, so this
// is an allowlist against remote keystroke injection, not a convenience.
var Allowed = []string{"model", "compact", "clear", "context", "cost", "status"}

// Validate reports whether text is a slash command whose name is on the
// allowlist, e.g. "/model sonnet". It ignores any arguments after the name.
func Validate(text string) bool {
	rest, ok := strings.CutPrefix(text, "/")
	if !ok {
		return false
	}

	name, _, _ := strings.Cut(rest, " ")

	for _, allowed := range Allowed {
		if name == allowed {
			return true
		}
	}

	return false
}
