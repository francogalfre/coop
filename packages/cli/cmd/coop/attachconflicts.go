package main

import (
	"fmt"
	"slices"

	"github.com/francogalfre/coop/packages/cli/internal/dirregistry"
	"github.com/francogalfre/coop/packages/cli/internal/harness"
)

// rejectConflictingAttach fails when an active attach already holds a harness
// that cannot multiplex coop attaches in one directory (opencode and pi share
// a single plugin/extension file, so a second attach would silently steal its
// events).
func rejectConflictingAttach(dir string, adapters []harness.Adapter) error {
	active, err := dirregistry.Load(dir)
	if err != nil {
		return fmt.Errorf("check active coop sessions in %s: %w", dir, err)
	}

	return checkConflict(adapters, active)
}

// checkConflict returns an error if any active entry uses a harness that does
// not support multiple attaches per directory and is also requested by adapters.
func checkConflict(adapters []harness.Adapter, active []dirregistry.Entry) error {
	names := adapterNames(adapters)

	for _, entry := range active {
		for _, name := range entry.Harnesses {
			if !slices.Contains(names, name) {
				continue
			}

			a, ok := harness.ByName(name, allAdapters)
			if ok && !a.SupportsMultiAttach() {
				return fmt.Errorf(
					"coop attach: %s already has an active coop session in this directory (session %s, pid %d); "+
						"%s supports only one coop session per directory, stop it first (Ctrl-C, then run `coop detach` to clean up)",
					name, entry.SessionID, entry.PID, name,
				)
			}
		}
	}

	return nil
}

func adapterNames(adapters []harness.Adapter) []string {
	names := make([]string, len(adapters))

	for i, a := range adapters {
		names[i] = a.Name()
	}

	return names
}
