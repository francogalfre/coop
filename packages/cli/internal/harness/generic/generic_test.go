package generic

import "testing"

func TestAdapterInstallIsNoop(t *testing.T) {
	dir := t.TempDir()

	inst, err := Adapter{}.Install(dir, "http://127.0.0.1:8788")
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

func TestAdapterDetectAlwaysTrue(t *testing.T) {
	if !(Adapter{}).Detect(t.TempDir()) {
		t.Fatal("Adapter.Detect() = false, want always true")
	}
}

func TestAdapterRemoveAllIsNoop(t *testing.T) {
	if err := (Adapter{}).RemoveAll(t.TempDir()); err != nil {
		t.Fatalf("RemoveAll() error = %v, want nil", err)
	}
}
