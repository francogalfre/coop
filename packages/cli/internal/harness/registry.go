package harness

// All returns every adapter in priority order, generic always last as the
// fallback (pty-wrap only, no hook events).
func All() []Adapter {
	return []Adapter{
		claudeCodeAdapter{},
		opencodeAdapter{},
		piAdapter{},
		genericAdapter{},
	}
}

// Detect returns every non-generic adapter usable in dir.
func Detect(dir string) []Adapter {
	var found []Adapter

	for _, a := range All() {
		if a.Name() == genericName {
			continue
		}
		if a.Detect(dir) {
			found = append(found, a)
		}
	}

	return found
}
