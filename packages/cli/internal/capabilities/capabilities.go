package capabilities

const claudeCode = "claude-code"

type Capabilities struct {
	Steer    bool `json:"steer"`
	Block    bool `json:"block"`
	Commands bool `json:"commands"`
	PTY      bool `json:"pty"`
}

// ForAttach reports capabilities under `coop attach`, where coop never owns
// the harness's terminal: steering is always possible (every attach-capable
// harness has an injection primitive - see harnesses.md), but only Claude
// Code's PreToolUse hook is wired as a blocking veto today; opencode and pi
// have their own veto hooks that coop does not wire yet.
func ForAttach(harnessName string) Capabilities {
	return Capabilities{
		Steer: true,
		Block: harnessName == claudeCode,
	}
}

// ForRun reports capabilities under `coop run`, where coop owns the pty:
// steering and slash commands are both plain keystroke injection regardless
// of harness, but there is no hook-deny mechanism wired for a blocking veto
// (the hook server's steer flag is false in run mode, so it never answers a
// PreToolUse hook with a deny - see hooks.Server).
func ForRun() Capabilities {
	return Capabilities{
		Steer:    true,
		Commands: true,
		PTY:      true,
	}
}
