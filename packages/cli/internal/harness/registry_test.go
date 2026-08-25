package harness

import "testing"

func TestAllReturnsAdaptersInPriorityOrderGenericLast(t *testing.T) {
	all := All()

	wantNames := []string{"claude-code", "opencode", "pi", "other"}
	if len(all) != len(wantNames) {
		t.Fatalf("got %d adapters, want %d", len(all), len(wantNames))
	}

	for i, a := range all {
		if a.Name() != wantNames[i] {
			t.Errorf("adapter %d: got %q, want %q", i, a.Name(), wantNames[i])
		}
	}
}

func TestDetectExcludesGeneric(t *testing.T) {
	dir := t.TempDir()

	for _, a := range Detect(dir) {
		if a.Name() == genericName {
			t.Fatal("Detect() returned the generic fallback adapter, want it excluded")
		}
	}
}
