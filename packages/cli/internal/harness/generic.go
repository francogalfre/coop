package harness

const genericName = "other"

// genericAdapter is the fallback for harnesses coop doesn't have hook
// integration for: no config is written, so coverage is whatever `coop run`
// gets from wrapping the pty.
type genericAdapter struct{}

func (genericAdapter) Name() string { return genericName }

func (genericAdapter) Detect(dir string) bool { return true }

func (genericAdapter) Install(dir, baseURL string) (Installation, error) {
	return Installation{Remove: func() error { return nil }}, nil
}
