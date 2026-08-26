package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/francogalfre/coop/packages/cli/internal/harness"
	"github.com/francogalfre/coop/packages/cli/internal/harness/claudecode"
	"github.com/francogalfre/coop/packages/cli/internal/harness/generic"
	"github.com/francogalfre/coop/packages/cli/internal/harness/opencode"
	"github.com/francogalfre/coop/packages/cli/internal/harness/pi"
)

const piHarnessName = pi.Name

var allAdapters = []harness.Adapter{
	claudecode.Adapter{},
	opencode.Adapter{},
	pi.Adapter{},
	generic.Adapter{},
}

func selectHarnesses(dir, harnessFlag string) ([]harness.Adapter, error) {
	if harnessFlag != "" {
		a, ok := harness.ByName(harnessFlag, allAdapters)
		if !ok {
			return nil, fmt.Errorf("harness: unknown --harness %q", harnessFlag)
		}
		return []harness.Adapter{a}, nil
	}

	detected := harness.Detect(dir, allAdapters)
	if len(detected) > 1 {
		names := make([]string, len(detected))
		for i, a := range detected {
			names[i] = a.Name()
		}
		return nil, fmt.Errorf("harness: multiple harnesses detected (%s), pass --harness=<name> to pick one", strings.Join(names, ", "))
	}

	return detected, nil
}

func selectRunHarness(harnessFlag, invokedName string) ([]harness.Adapter, error) {
	if harnessFlag != "" {
		a, ok := harness.ByName(harnessFlag, allAdapters)
		if !ok {
			return nil, fmt.Errorf("harness: unknown --harness %q", harnessFlag)
		}
		return []harness.Adapter{a}, nil
	}

	if a, ok := harness.ByBinary(filepath.Base(invokedName), allAdapters); ok {
		return []harness.Adapter{a}, nil
	}

	return nil, nil
}

func installHarnesses(dir, baseURL string, adapters []harness.Adapter) ([]harness.Installation, error) {
	if len(adapters) == 0 {
		fmt.Println("coop: no supported harness detected (hook events disabled)")
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
