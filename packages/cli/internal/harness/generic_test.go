package harness

import "testing"

func TestGenericAdapterInstallIsNoop(t *testing.T) {
	dir := t.TempDir()

	inst, err := genericAdapter{}.Install(dir, "http://127.0.0.1:8788")
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if len(inst.Paths) != 0 {
		t.Fatalf("got Paths %v, want none", inst.Paths)
	}

	if err := inst.Remove(); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
}

func TestGenericAdapterDetectAlwaysTrue(t *testing.T) {
	if !(genericAdapter{}).Detect(t.TempDir()) {
		t.Fatal("genericAdapter.Detect() = false, want always true")
	}
}
