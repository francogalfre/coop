package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallRemoveRemovesOnlyThisAttachEntry(t *testing.T) {
	dir := t.TempDir()

	instA, err := (Adapter{}).Install(dir, "http://127.0.0.1:11111")
	if err != nil {
		t.Fatalf("install first attach: %v", err)
	}

	instB, err := (Adapter{}).Install(dir, "http://127.0.0.1:22222")
	if err != nil {
		t.Fatalf("install second attach: %v", err)
	}

	if err := instA.Remove(); err != nil {
		t.Fatalf("first attach Remove(): %v", err)
	}

	root := readSettings(t, settingsPath(dir))

	var hooksRoot map[string][]hookGroup
	if err := json.Unmarshal(root["hooks"], &hooksRoot); err != nil {
		t.Fatalf("unmarshal hooks: %v", err)
	}

	for _, event := range hookEventNames {
		urls := coopEntryURLs(hooksRoot[event])

		want := "http://127.0.0.1:22222/hook/claude-code/" + event
		if len(urls) != 1 || urls[0] != want {
			t.Fatalf("event %s: coop entries %v, want only the second attach's %q", event, urls, want)
		}
	}

	if err := instB.Remove(); err != nil {
		t.Fatalf("second attach Remove(): %v", err)
	}

	if _, err := os.Stat(settingsPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("settings file still exists after both attaches removed, err = %v", err)
	}
}

func TestRemoveAllRemovesEveryAttachEntry(t *testing.T) {
	dir := t.TempDir()

	if _, err := (Adapter{}).Install(dir, "http://127.0.0.1:11111"); err != nil {
		t.Fatalf("install first attach: %v", err)
	}
	if _, err := (Adapter{}).Install(dir, "http://127.0.0.1:22222"); err != nil {
		t.Fatalf("install second attach: %v", err)
	}

	if err := (Adapter{}).RemoveAll(dir); err != nil {
		t.Fatalf("RemoveAll(): %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, settingsDir, settingsFileName)); !os.IsNotExist(err) {
		t.Fatalf("settings file still exists after RemoveAll, err = %v", err)
	}
}
