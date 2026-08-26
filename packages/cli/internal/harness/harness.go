package harness

type Event struct {
	Body []byte
}

type Steer struct {
	From, Text string
}

type Adapter interface {
	Name() string
	Binary() string
	IsFallback() bool
	Detect(dir string) bool
	Install(dir, baseURL string) (Installation, error)
	RemoveAll(dir string) error
}

type Installation struct {
	Paths  []string
	Remove func() error
}
