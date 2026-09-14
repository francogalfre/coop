package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/francogalfre/coop/packages/cli/internal/dirregistry"
	"github.com/francogalfre/coop/packages/cli/internal/harness"
)

func claimDirectory(dir, sessionID, baseURL string, adapters []harness.Adapter, install func() error) error {
	entry := dirregistry.Entry{
		SessionID: sessionID,
		PID:       os.Getpid(),
		BaseURL:   baseURL,
		Harnesses: adapterNames(adapters),
	}

	return dirregistry.Claim(dir, entry, func(active []dirregistry.Entry) error {
		return checkConflict(adapters, active)
	}, install)
}

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
