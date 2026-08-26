package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type CLICredentials struct {
	Token       string `json:"token"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func CredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: user home dir: %w", err)
	}

	return filepath.Join(home, ".config", "coop", "credentials.json"), nil
}

func SaveCredentials(cred CLICredentials) error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("config: create credentials dir: %w", err)
	}

	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("config: marshal credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("config: write credentials: %w", err)
	}

	return nil
}

func loadCredentials() (CLICredentials, error) {
	path, err := CredentialsPath()
	if err != nil {
		return CLICredentials{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CLICredentials{}, nil
		}

		return CLICredentials{}, fmt.Errorf("config: read credentials: %w", err)
	}

	var cred CLICredentials

	if err := json.Unmarshal(data, &cred); err != nil {
		return CLICredentials{}, fmt.Errorf("config: parse credentials: %w", err)
	}

	return cred, nil
}
