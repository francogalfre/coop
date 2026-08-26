package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/relayclient"
)

const (
	githubDeviceCodeURL      = "https://github.com/login/device/code"
	githubAccessTokenURL     = "https://github.com/login/oauth/access_token"
	githubDeviceCodeScope    = "read:user user:email"
	deviceFlowRequestTimeout = 10 * time.Second
)

var deviceFlowHTTPClient = &http.Client{Timeout: deviceFlowRequestTimeout}

type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

func runLogin(ctx context.Context, cfg config.Config) error {
	clientID := strings.TrimSpace(os.Getenv("COOP_GITHUB_CLIENT_ID"))
	if clientID == "" {
		return fmt.Errorf("COOP_GITHUB_CLIENT_ID is not set: the GitHub OAuth App's public client id is required for coop login")
	}

	device, err := requestDeviceCode(ctx, clientID)
	if err != nil {
		return fmt.Errorf("coop login: request device code: %w", err)
	}

	fmt.Printf("Go to %s and enter code: %s\n", device.VerificationURI, device.UserCode)

	githubAccessToken, err := pollForAccessToken(ctx, clientID, device)
	if err != nil {
		return fmt.Errorf("coop login: %w", err)
	}

	result, err := relayclient.Login(ctx, cfg, githubAccessToken)
	if err != nil {
		return fmt.Errorf("coop login: exchange token with relay: %w", err)
	}

	if err := config.SaveCredentials(config.CLICredentials{
		Token:       result.Token,
		Username:    result.Username,
		DisplayName: result.DisplayName,
	}); err != nil {
		return fmt.Errorf("coop login: save credentials: %w", err)
	}

	fmt.Printf("Logged in as %s\n", result.Username)

	return nil
}

func requestDeviceCode(ctx context.Context, clientID string) (deviceCodeResponse, error) {
	form := url.Values{
		"client_id": {clientID},
		"scope":     {githubDeviceCodeScope},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubDeviceCodeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return deviceCodeResponse{}, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := deviceFlowHTTPClient.Do(req)
	if err != nil {
		return deviceCodeResponse{}, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return deviceCodeResponse{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var device deviceCodeResponse

	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return deviceCodeResponse{}, fmt.Errorf("decode response: %w", err)
	}

	return device, nil
}

func pollForAccessToken(ctx context.Context, clientID string, device deviceCodeResponse) (string, error) {
	interval := time.Duration(device.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	deadline := time.Now().Add(time.Duration(device.ExpiresIn) * time.Second)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return "", fmt.Errorf("device code expired before login was completed")
			}

			token, tokenErr, err := requestAccessToken(ctx, clientID, device.DeviceCode)
			if err != nil {
				return "", err
			}

			switch tokenErr {
			case "":
				return token, nil
			case "authorization_pending":
				continue
			case "slow_down":
				interval += 5 * time.Second
				ticker.Reset(interval)
				continue
			default:
				return "", fmt.Errorf("github device flow error: %s", tokenErr)
			}
		}
	}
}

func requestAccessToken(ctx context.Context, clientID, deviceCode string) (token, tokenErr string, err error) {
	form := url.Values{
		"client_id":   {clientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubAccessTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := deviceFlowHTTPClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result accessTokenResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}

	return result.AccessToken, result.Error, nil
}
