package config

import "testing"

func TestLoadDefaultsAddr(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/coop_test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != defaultAddr {
		t.Fatalf("got %q, want default %q", cfg.Addr, defaultAddr)
	}
}

func TestLoadReadsAddrFromEnv(t *testing.T) {
	t.Setenv("COOP_RELAY_ADDR", ":9999")
	t.Setenv("DATABASE_URL", "postgres://localhost/coop_test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != ":9999" {
		t.Fatalf("got %q, want :9999", cfg.Addr)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load: want error when DATABASE_URL is unset")
	}
}
