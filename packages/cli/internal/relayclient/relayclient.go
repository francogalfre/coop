package relayclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

const requestTimeout = 10 * time.Second

var httpClient = &http.Client{Timeout: requestTimeout}

func PostEvent(ctx context.Context, cfg config.Config, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.RelayURL+"/v1/events", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("relayclient: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	setCLICredential(req, cfg)
	setProject(req, cfg)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("relayclient: post event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("relayclient: post event: unexpected status %d: %s", resp.StatusCode, readBody(resp.Body))
	}

	return nil
}

type TakeoverInfo struct {
	Active bool
	By     string
}

type SteerResult struct {
	From       string
	Text       string
	HasMessage bool
	Takeover   TakeoverInfo
}

type steerGetBody struct {
	HasMessage bool   `json:"has_message"`
	From       string `json:"from"`
	Text       string `json:"text"`
	Takeover   struct {
		Active bool   `json:"active"`
		By     string `json:"by"`
	} `json:"takeover"`
}

func GetSteer(ctx context.Context, cfg config.Config, sessionID string) (SteerResult, error) {
	steerURL := cfg.RelayURL + "/v1/sessions/" + url.PathEscape(sessionID) + "/steer"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, steerURL, nil)
	if err != nil {
		return SteerResult{}, fmt.Errorf("relayclient: build request: %w", err)
	}

	setCLICredential(req, cfg)

	resp, err := httpClient.Do(req)
	if err != nil {
		return SteerResult{}, fmt.Errorf("relayclient: get steer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SteerResult{}, fmt.Errorf("relayclient: get steer: unexpected status %d: %s", resp.StatusCode, readBody(resp.Body))
	}

	var body steerGetBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return SteerResult{}, fmt.Errorf("relayclient: decode steer response: %w", err)
	}

	return SteerResult{
		From:       body.From,
		Text:       body.Text,
		HasMessage: body.HasMessage,
		Takeover:   TakeoverInfo{Active: body.Takeover.Active, By: body.Takeover.By},
	}, nil
}

type LoginResult struct {
	Token       string
	Username    string
	DisplayName string
	AvatarURL   string
}

func Login(ctx context.Context, cfg config.Config, githubAccessToken string) (LoginResult, error) {
	payload, err := json.Marshal(map[string]string{"github_access_token": githubAccessToken})
	if err != nil {
		return LoginResult{}, fmt.Errorf("relayclient: login: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.RelayURL+"/v1/auth/cli/exchange", bytes.NewReader(payload))
	if err != nil {
		return LoginResult{}, fmt.Errorf("relayclient: login: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return LoginResult{}, fmt.Errorf("relayclient: login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return LoginResult{}, fmt.Errorf("relayclient: login: unexpected status %d: %s", resp.StatusCode, readBody(resp.Body))
	}

	var result struct {
		Token       string `json:"token"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return LoginResult{}, fmt.Errorf("relayclient: login: decode response: %w", err)
	}

	return LoginResult{
		Token:       result.Token,
		Username:    result.Username,
		DisplayName: result.DisplayName,
		AvatarURL:   result.AvatarURL,
	}, nil
}

func setCLICredential(req *http.Request, cfg config.Config) {
	if cfg.CLICredential != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.CLICredential)
	}
}

func setProject(req *http.Request, cfg config.Config) {
	if cfg.Project != "" {
		req.Header.Set("X-Coop-Project", cfg.Project)
	}
}

func readBody(r io.Reader) string {
	b, err := io.ReadAll(io.LimitReader(r, 4096))
	if err != nil {
		return ""
	}

	return string(b)
}
