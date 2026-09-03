package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	RelayURL      string
	SessionID     string
	HookAddr      string
	SessionToken  string
	CLICredential string
	Username      string
	DisplayName   string
	Project       string
	// DisableDiffStreaming defaults false (streaming on) so every existing
	// Config{} zero value keeps today's behavior; COOP_DISABLE_STREAM_DIFFS
	// opts out.
	DisableDiffStreaming bool
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

	sessionToken := os.Getenv("COOP_SESSION_TOKEN")
	if sessionToken == "" {
		generated, err := generateHexToken(32)
		if err != nil {
			return Config{}, fmt.Errorf("config: generate session token: %w", err)
		}

		sessionToken = generated
	}

	cred, err := loadCredentials()
	if err != nil {
		return Config{}, err
	}

	disableDiffStreaming, _ := strconv.ParseBool(os.Getenv("COOP_DISABLE_STREAM_DIFFS"))

	return Config{
		RelayURL:             relayURL,
		SessionID:            sessionID,
		HookAddr:             hookAddr,
		SessionToken:         sessionToken,
		CLICredential:        cred.Token,
		Username:             cred.Username,
		DisplayName:          cred.DisplayName,
		DisableDiffStreaming: disableDiffStreaming,
	}, nil
}

func generateSessionID() (string, error) {
	token, err := generateHexToken(16)
	if err != nil {
		return "", err
	}

	return "sess-" + token, nil
}

func generateHexToken(byteLen int) (string, error) {
	buf := make([]byte, byteLen)

	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return hex.EncodeToString(buf), nil
}
