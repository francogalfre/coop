package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

type Config struct {
	RelayURL  string
	SessionID string
	HookAddr  string
}

const (
	defaultRelayURL = "http://localhost:8787"
	defaultHookAddr = "127.0.0.1:8788"
)

func Load() (Config, error) {
	relayURL := os.Getenv("COOP_RELAY_URL")
	if relayURL == "" {
		relayURL = defaultRelayURL
	}

	hookAddr := os.Getenv("COOP_HOOK_ADDR")
	if hookAddr == "" {
		hookAddr = defaultHookAddr
	}

	sessionID := os.Getenv("COOP_SESSION_ID")
	if sessionID == "" {
		generated, err := generateSessionID()
		if err != nil {
			return Config{}, fmt.Errorf("config: generate session id: %w", err)
		}

		sessionID = generated
	}

	return Config{RelayURL: relayURL, SessionID: sessionID, HookAddr: hookAddr}, nil
}

func generateSessionID() (string, error) {
	buf := make([]byte, 16)

	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return "sess-" + hex.EncodeToString(buf), nil
}
