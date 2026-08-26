package harness

import "errors"

func RemoveAllTraces(dir string, adapters []Adapter) error {
	errs := make([]error, len(adapters))

	for i, a := range adapters {
		errs[i] = a.RemoveAll(dir)
	}

	return errors.Join(errs...)
}
