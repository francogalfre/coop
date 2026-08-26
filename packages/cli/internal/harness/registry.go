package harness

func Detect(dir string, adapters []Adapter) []Adapter {
	var found []Adapter

	for _, a := range adapters {
		if a.IsFallback() {
			continue
		}
		if a.Detect(dir) {
			found = append(found, a)
		}
	}

	return found
}

func ByName(name string, adapters []Adapter) (Adapter, bool) {
	for _, a := range adapters {
		if a.Name() == name {
			return a, true
		}
	}

	return nil, false
}

func ByBinary(name string, adapters []Adapter) (Adapter, bool) {
	for _, a := range adapters {
		if !a.IsFallback() && a.Binary() == name {
			return a, true
		}
	}

	return nil, false
}
