package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveCredentialsThenLoadRoundTrips(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	want := CLICredentials{
		Token:       "deadbeef",
		UserID:      "user-123",
		Username:    "octocat",
		DisplayName: "The Octocat",
		AvatarURL:   "https://example.com/avatar.png",
	}

	if err := SaveCredentials(want); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	got, err := loadCredentials()
	if err != nil {
		t.Fatalf("loadCredentials: %v", err)
	}

	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestSaveCredentialsWritesRestrictivePermissions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := SaveCredentials(CLICredentials{Token: "deadbeef"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	path, err := CredentialsPath()
	if err != nil {
		t.Fatalf("CredentialsPath: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("got file mode %o, want %o", perm, 0o600)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}

	if perm := dirInfo.Mode().Perm(); perm != 0o700 {
		t.Fatalf("got dir mode %o, want %o", perm, 0o700)
	}
}

func TestLoadCredentialsReturnsZeroValueWhenAbsent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got, err := loadCredentials()
	if err != nil {
		t.Fatalf("loadCredentials: %v", err)
	}

	if got != (CLICredentials{}) {
		t.Fatalf("got %+v, want zero value", got)
	}
}
