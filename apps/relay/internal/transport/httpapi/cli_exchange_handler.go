package httpapi

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
)

const webResolveTimeout = 10 * time.Second

var webResolveClient = &http.Client{Timeout: webResolveTimeout}

type cliExchangeRequestBody struct {
	GitHubAccessToken string `json:"github_access_token"`
}

type cliExchangeResponseBody struct {
	Token       string `json:"token"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type resolvedGitHubUser struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
}

func handleCLIExchange(cfg config.Config, pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body cliExchangeRequestBody

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.GitHubAccessToken == "" {
			writeError(w, http.StatusBadRequest, "github_access_token: required")
			return
		}

		resolved, err := resolveGitHubUser(r.Context(), cfg, body.GitHubAccessToken)
		if err != nil {
			writeError(w, http.StatusBadGateway, fmt.Sprintf("resolve github user: %v", err))
			return
		}

		rawToken, err := pool.CreateCliCredential(r.Context(), resolved.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create cli credential")
			return
		}

		writeJSON(w, http.StatusOK, cliExchangeResponseBody{
			Token:       hex.EncodeToString(rawToken),
			Username:    resolved.Username,
			DisplayName: resolved.DisplayName,
			AvatarURL:   resolved.AvatarURL,
		})
	}
}

func resolveGitHubUser(ctx context.Context, cfg config.Config, githubAccessToken string) (resolvedGitHubUser, error) {
	payload, err := json.Marshal(map[string]string{"githubAccessToken": githubAccessToken})
	if err != nil {
		return resolvedGitHubUser{}, fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.WebInternalURL+"/api/internal/users/resolve-github", bytes.NewReader(payload))
	if err != nil {
		return resolvedGitHubUser{}, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Coop-Internal-Secret", cfg.InternalSecret)

	resp, err := webResolveClient.Do(req)
	if err != nil {
		return resolvedGitHubUser{}, fmt.Errorf("request web app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resolvedGitHubUser{}, fmt.Errorf("web app returned status %d: %s", resp.StatusCode, readLimited(resp.Body))
	}

	var resolved resolvedGitHubUser

	if err := json.NewDecoder(resp.Body).Decode(&resolved); err != nil {
		return resolvedGitHubUser{}, fmt.Errorf("decode response: %w", err)
	}

	return resolved, nil
}

func readLimited(r io.Reader) string {
	b, err := io.ReadAll(io.LimitReader(r, 4096))
	if err != nil {
		return ""
	}

	return string(b)
}
