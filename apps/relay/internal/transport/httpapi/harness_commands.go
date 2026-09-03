package httpapi

// Mirrors packages/protocol/src/shared/commands.ts's HARNESS_COMMANDS — an allowlist, since this becomes keystrokes in someone else's terminal.
var harnessCommands = map[string]bool{
	"model":   true,
	"compact": true,
	"clear":   true,
	"context": true,
	"cost":    true,
	"status":  true,
}
