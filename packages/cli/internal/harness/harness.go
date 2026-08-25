// Package harness detects which coding agent harness (Claude Code, opencode,
// pi, or none of the above) is usable in a project directory, and installs
// the harness-specific wiring that gets its events flowing into coop's hook
// server.
package harness

// Event mirrors one full protocol event JSON body, the same shape
// internal/hooks/envelope.go already produces.
type Event struct {
	Body []byte
}

// Steer is an attributed message a teammate sent through the relay.
type Steer struct {
	From, Text string
}

type Adapter interface {
	// Name becomes session.start.harness on the wire.
	Name() string
	// Detect reports whether this harness is usable in dir.
	Detect(dir string) bool
	// Install wires the harness in dir to POST hook events at baseURL.
	Install(dir, baseURL string) (Installation, error)
}

type Installation struct {
	Paths  []string
	Remove func() error
}
