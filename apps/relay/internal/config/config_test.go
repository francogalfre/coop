package config

import "testing"

func TestLoadDefaultsAddr(t *testing.T) {
	cfg := Load()
	if cfg.Addr != defaultAddr {
		t.Fatalf("got %q, want default %q", cfg.Addr, defaultAddr)
	}
}

func TestLoadReadsAddrFromEnv(t *testing.T) {
	t.Setenv("COOP_RELAY_ADDR", ":9999")

	cfg := Load()
	if cfg.Addr != ":9999" {
		t.Fatalf("got %q, want :9999", cfg.Addr)
	}
}
