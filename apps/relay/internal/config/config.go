package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Addr        string
	DatabaseURL string
	WebOrigins  []string

	GitHubClientID     string
	GitHubClientSecret string

	InternalSecret string
	WebInternalURL string
}

const (
	defaultAddr       = ":8787"
	defaultWebOrigin  = "http://localhost:3000"
	webOriginsEnvName = "COOP_WEB_ORIGINS"
)

func Load() (Config, error) {
	addr := os.Getenv("COOP_RELAY_ADDR")

	if addr == "" {
		addr = defaultAddr
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}

	webOrigins := []string{defaultWebOrigin}
	if raw := os.Getenv(webOriginsEnvName); raw != "" {
		webOrigins = splitAndTrim(raw)
	}

	return Config{
		Addr:               addr,
		DatabaseURL:        databaseURL,
		WebOrigins:         webOrigins,
		GitHubClientID:     os.Getenv("COOP_GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("COOP_GITHUB_CLIENT_SECRET"),
		InternalSecret:     os.Getenv("COOP_INTERNAL_SECRET"),
		WebInternalURL:     os.Getenv("COOP_WEB_INTERNAL_URL"),
	}, nil
}

func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}
