package config

import "testing"

func TestLoadDefaultsRelayURL(t *testing.T) {
	t.Setenv("COOP_RELAY_URL", "")
	t.Setenv("COOP_SESSION_ID", "sess-fixed")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.RelayURL != defaultRelayURL {
		t.Fatalf("got %q, want default %q", cfg.RelayURL, defaultRelayURL)
	}
}

func TestLoadReadsRelayURLFromEnv(t *testing.T) {
	t.Setenv("COOP_RELAY_URL", "http://relay.example:9000")
	t.Setenv("COOP_SESSION_ID", "sess-fixed")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.RelayURL != "http://relay.example:9000" {
		t.Fatalf("got %q, want http://relay.example:9000", cfg.RelayURL)
	}
}

func TestLoadReadsSessionIDFromEnv(t *testing.T) {
	t.Setenv("COOP_SESSION_ID", "sess-explicit")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SessionID != "sess-explicit" {
		t.Fatalf("got %q, want sess-explicit", cfg.SessionID)
	}
}

func TestLoadGeneratesSessionIDWhenUnset(t *testing.T) {
	t.Setenv("COOP_SESSION_ID", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SessionID == "" {
		t.Fatal("got empty session id, want a generated one")
	}
}

func TestLoadDefaultsHookAddr(t *testing.T) {
	t.Setenv("COOP_HOOK_ADDR", "")
	t.Setenv("COOP_SESSION_ID", "sess-fixed")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HookAddr != defaultHookAddr {
		t.Fatalf("got %q, want default %q", cfg.HookAddr, defaultHookAddr)
	}
}

func TestLoadReadsHookAddrFromEnv(t *testing.T) {
	t.Setenv("COOP_HOOK_ADDR", "127.0.0.1:9999")
	t.Setenv("COOP_SESSION_ID", "sess-fixed")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HookAddr != "127.0.0.1:9999" {
		t.Fatalf("got %q, want 127.0.0.1:9999", cfg.HookAddr)
	}
}

func TestLoadPopulatesCLICredentialFromSavedFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("COOP_SESSION_ID", "sess-fixed")

	if err := SaveCredentials(CLICredentials{Token: "deadbeef", Username: "octocat"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CLICredential != "deadbeef" {
		t.Fatalf("got %q, want %q", cfg.CLICredential, "deadbeef")
	}
}

func TestLoadPopulatesIdentityFromSavedCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("COOP_SESSION_ID", "sess-fixed")

	saved := CLICredentials{
		Token:       "deadbeef",
		UserID:      "user-123",
		Username:    "octocat",
		DisplayName: "The Octocat",
		AvatarURL:   "https://example.com/avatar.png",
	}
	if err := SaveCredentials(saved); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.UserID != "user-123" || cfg.Username != "octocat" || cfg.DisplayName != "The Octocat" || cfg.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("got %+v, want identity from saved credentials", cfg)
	}
}

func TestLoadLeavesCLICredentialEmptyWhenNoFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("COOP_SESSION_ID", "sess-fixed")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CLICredential != "" {
		t.Fatalf("got %q, want empty", cfg.CLICredential)
	}
}

func TestLoadDefaultsDiffStreamingOn(t *testing.T) {
	t.Setenv("COOP_DISABLE_STREAM_DIFFS", "")
	t.Setenv("COOP_SESSION_ID", "sess-fixed")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DisableDiffStreaming {
		t.Fatal("got DisableDiffStreaming = true, want false (diff streaming is on by default)")
	}
}

func TestLoadReadsDisableStreamDiffsFromEnv(t *testing.T) {
	t.Setenv("COOP_DISABLE_STREAM_DIFFS", "true")
	t.Setenv("COOP_SESSION_ID", "sess-fixed")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.DisableDiffStreaming {
		t.Fatal("got DisableDiffStreaming = false, want true")
	}
}

func TestLoadGeneratesUniqueSessionIDs(t *testing.T) {
	t.Setenv("COOP_SESSION_ID", "")

	first, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	second, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if first.SessionID == second.SessionID {
		t.Fatalf("got the same session id twice: %q", first.SessionID)
	}
}
