package generic

import "github.com/francogalfre/coop/packages/cli/internal/harness"

const Name = "other"

type Adapter struct{}

func (Adapter) Name() string { return Name }

func (Adapter) Binary() string { return "" }

func (Adapter) IsFallback() bool { return true }

func (Adapter) Detect(dir string) bool { return true }

func (Adapter) Install(dir, baseURL string) (harness.Installation, error) {
	return harness.Installation{Remove: func() error { return nil }}, nil
}

func (Adapter) RemoveAll(dir string) error { return nil }
