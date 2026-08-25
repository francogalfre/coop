package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/francogalfre/coop/packages/cli/internal/harness"
)

const piHarnessName = "pi"

// installHarnesses wires every harness detected in dir up to baseURL. On a
// partial failure it unwinds whatever it already installed before
// returning the error.
func installHarnesses(dir, baseURL string) ([]harness.Installation, error) {
	adapters := harness.Detect(dir)
	if len(adapters) == 0 {
		fmt.Println("coop: no supported harness detected in this directory (hook events disabled)")
		return nil, nil
	}

	installations := make([]harness.Installation, 0, len(adapters))

	for _, a := range adapters {
		inst, err := a.Install(dir, baseURL)
		if err != nil {
			removeInstallations(installations)
			return nil, fmt.Errorf("harness: install %s: %w", a.Name(), err)
		}

		installations = append(installations, inst)
		printInstalled(a, inst)
	}

	return installations, nil
}

func printInstalled(a harness.Adapter, inst harness.Installation) {
	fmt.Printf("coop: %s detected", a.Name())
	if len(inst.Paths) > 0 {
		fmt.Printf(", installed at %s", strings.Join(inst.Paths, ", "))
	}
	fmt.Println()

	if a.Name() == piHarnessName {
		fmt.Println("coop: pi requires you to trust this project interactively before its extensions load")
	}
}

func removeInstallations(installations []harness.Installation) {
	for _, inst := range installations {
		if inst.Remove == nil {
			continue
		}
		if err := inst.Remove(); err != nil {
			fmt.Fprintf(os.Stderr, "coop: cleanup: %v\n", err)
		}
	}
}
