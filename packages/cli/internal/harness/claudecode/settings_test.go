package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readSettings(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}

	return root
}

func TestWriteSettingsCreatesFileWhenAbsent(t *testing.T) {
	dir := t.TempDir()

	path, err := writeSettings(dir, "http://127.0.0.1:12345")
	if err != nil {
		t.Fatalf("writeSettings() error = %v", err)
	}

	if path != filepath.Join(dir, settingsDir, settingsFileName) {
		t.Fatalf("got path %q", path)
	}

	root := readSettings(t, path)

	var hooksRoot map[string][]hookGroup
	if err := json.Unmarshal(root["hooks"], &hooksRoot); err != nil {
		t.Fatalf("unmarshal hooks: %v", err)
	}

	for _, event := range hookEventNames {
		groups, ok := hooksRoot[event]
		if !ok || len(groups) != 1 || len(groups[0].Hooks) != 1 {
			t.Fatalf("event %s: got groups %+v, want exactly one http hook entry", event, groups)
		}

		entry := groups[0].Hooks[0]
		wantURL := "http://127.0.0.1:12345/hook/claude-code/" + event
		if entry.Type != "http" || entry.URL != wantURL {
			t.Fatalf("event %s: got entry %+v, want type=http url=%s", event, entry, wantURL)
		}
	}
}

func TestWriteSettingsPreservesUnrelatedTopLevelKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, settingsDir, settingsFileName)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"permissions":{"allow":["Bash(ls:*)"]}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := writeSettings(dir, "http://127.0.0.1:12345")
	if err != nil {
		t.Fatalf("writeSettings() error = %v", err)
	}

	root := readSettings(t, got)
	if _, ok := root["permissions"]; !ok {
		t.Fatal("permissions key was dropped, want it preserved")
	}
}

func TestWriteSettingsPreservesOtherToolsHookEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, settingsDir, settingsFileName)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	existing := `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[` +
		`{"type":"command","command":"echo other-tool","timeout":30}]}]}}`
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := writeSettings(dir, "http://127.0.0.1:12345")
	if err != nil {
		t.Fatalf("writeSettings() error = %v", err)
	}

	root := readSettings(t, got)

	var hooksRoot map[string]json.RawMessage
	if err := json.Unmarshal(root["hooks"], &hooksRoot); err != nil {
		t.Fatalf("unmarshal hooks: %v", err)
	}

	var rawGroups []map[string]json.RawMessage
	if err := json.Unmarshal(hooksRoot["PreToolUse"], &rawGroups); err != nil {
		t.Fatalf("unmarshal PreToolUse groups: %v", err)
	}

	if len(rawGroups) != 2 {
		t.Fatalf("got %d groups for PreToolUse, want 2 (other tool's + coop's)", len(rawGroups))
	}

	var foreignGroup map[string]json.RawMessage
	for _, g := range rawGroups {
		var matcher string
		if err := json.Unmarshal(g["matcher"], &matcher); err == nil && matcher == "Bash" {
			foreignGroup = g
		}
	}

	if foreignGroup == nil {
		t.Fatal("foreign tool's group (matcher=Bash) was dropped")
	}

	var matcher string
	if err := json.Unmarshal(foreignGroup["matcher"], &matcher); err != nil || matcher != "Bash" {
		t.Fatalf("got matcher %v, want \"Bash\" preserved byte-for-byte", foreignGroup["matcher"])
	}

	var foreignHooks []map[string]json.RawMessage
	if err := json.Unmarshal(foreignGroup["hooks"], &foreignHooks); err != nil {
		t.Fatalf("unmarshal foreign hooks: %v", err)
	}
	if len(foreignHooks) != 1 {
		t.Fatalf("got %d foreign hook entries, want 1", len(foreignHooks))
	}

	var command string
	if err := json.Unmarshal(foreignHooks[0]["command"], &command); err != nil || command != "echo other-tool" {
		t.Fatalf("got command %v, want %q preserved byte-for-byte", foreignHooks[0]["command"], "echo other-tool")
	}

	var timeout int
	if err := json.Unmarshal(foreignHooks[0]["timeout"], &timeout); err != nil || timeout != 30 {
		t.Fatalf("got timeout %v, want 30 preserved byte-for-byte", foreignHooks[0]["timeout"])
	}

	var typ string
	if err := json.Unmarshal(foreignHooks[0]["type"], &typ); err != nil || typ != "command" {
		t.Fatalf("got type %v, want \"command\" preserved", foreignHooks[0]["type"])
	}

	var hooksRootTyped map[string][]hookGroup
	if err := json.Unmarshal(root["hooks"], &hooksRootTyped); err != nil {
		t.Fatalf("unmarshal typed hooks: %v", err)
	}

	foundCoop := false
	for _, g := range hooksRootTyped["PreToolUse"] {
		for _, e := range g.Hooks {
			if isCoopEntry(e) {
				foundCoop = true
			}
		}
	}
	if !foundCoop {
		t.Error("coop's http hook entry was not written")
	}
}

func TestWriteSettingsReplacesPriorCoopEntryInsteadOfAccumulating(t *testing.T) {
	dir := t.TempDir()

	if _, err := writeSettings(dir, "http://127.0.0.1:11111"); err != nil {
		t.Fatalf("first writeSettings() error = %v", err)
	}

	path, err := writeSettings(dir, "http://127.0.0.1:22222")
	if err != nil {
		t.Fatalf("second writeSettings() error = %v", err)
	}

	root := readSettings(t, path)

	var hooksRoot map[string][]hookGroup
	if err := json.Unmarshal(root["hooks"], &hooksRoot); err != nil {
		t.Fatalf("unmarshal hooks: %v", err)
	}

	groups := hooksRoot["PreToolUse"]

	total := 0
	for _, g := range groups {
		total += len(g.Hooks)
	}

	if total != 1 {
		t.Fatalf("got %d hook entries for PreToolUse after two attach runs, want 1 (replaced, not accumulated)", total)
	}

	entry := groups[0].Hooks[0]
	if entry.URL != "http://127.0.0.1:22222/hook/claude-code/PreToolUse" {
		t.Fatalf("got url %q, want the second run's port", entry.URL)
	}
}

func TestWriteSettingsStartsFreshOnMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, settingsDir, settingsFileName)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{not valid json`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := writeSettings(dir, "http://127.0.0.1:12345")
	if err != nil {
		t.Fatalf("writeSettings() error = %v", err)
	}

	root := readSettings(t, got)
	if _, ok := root["hooks"]; !ok {
		t.Fatal("hooks key missing after recovering from malformed JSON")
	}
}
