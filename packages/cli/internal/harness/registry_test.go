package harness_test

import (
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/harness"
	"github.com/francogalfre/coop/packages/cli/internal/harness/claudecode"
	"github.com/francogalfre/coop/packages/cli/internal/harness/generic"
	"github.com/francogalfre/coop/packages/cli/internal/harness/opencode"
	"github.com/francogalfre/coop/packages/cli/internal/harness/pi"
)

func testAdapters() []harness.Adapter {
	return []harness.Adapter{claudecode.Adapter{}, opencode.Adapter{}, pi.Adapter{}, generic.Adapter{}}
}

func TestDetectExcludesGeneric(t *testing.T) {
	dir := t.TempDir()

	for _, a := range harness.Detect(dir, testAdapters()) {
		if a.Name() == generic.Name {
			t.Fatal("Detect() returned the generic fallback adapter, want it excluded")
		}
	}
}

func TestByNameFindsEachAdapter(t *testing.T) {
	adapters := testAdapters()

	for _, name := range []string{claudecode.Name, opencode.Name, pi.Name, generic.Name} {
		a, ok := harness.ByName(name, adapters)
		if !ok || a.Name() != name {
			t.Fatalf("ByName(%q) = %v, %v", name, a, ok)
		}
	}

	if _, ok := harness.ByName("unknown", adapters); ok {
		t.Fatal("ByName(\"unknown\") = true, want false")
	}
}

func TestByBinaryExcludesGeneric(t *testing.T) {
	adapters := testAdapters()

	a, ok := harness.ByBinary("claude", adapters)
	if !ok || a.Name() != claudecode.Name {
		t.Fatalf("ByBinary(\"claude\") = %v, %v", a, ok)
	}

	if _, ok := harness.ByBinary("", adapters); ok {
		t.Fatal("ByBinary(\"\") matched the generic fallback adapter, want it excluded")
	}
}
