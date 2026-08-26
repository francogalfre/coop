package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
)

const (
	browserSessionCookieName = "better-auth.session_token"
	webVerifyTimeout         = 10 * time.Second
)

var webVerifyClient = &http.Client{Timeout: webVerifyTimeout}

type sessionVerifyRequestBody struct {
	Cookie string `json:"cookie"`
}

type sessionVerifyResponseBody struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

func RequireBrowserSession(cfg config.Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(browserSessionCookieName)
			if err != nil || cookie.Value == "" {
				http.NotFound(w, r)
				return
			}

			actor, err := verifyBrowserSession(r, cfg, cookie.Value)
			if err != nil || actor.UserID == "" {
				http.NotFound(w, r)
				return
			}

			next(w, r.WithContext(WithActor(r.Context(), actor)))
		}
	}
}

func verifyBrowserSession(r *http.Request, cfg config.Config, cookieValue string) (Actor, error) {
	payload, err := json.Marshal(sessionVerifyRequestBody{Cookie: cookieValue})
	if err != nil {
		return Actor{}, fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, cfg.WebInternalURL+"/api/internal/session/verify", bytes.NewReader(payload))
	if err != nil {
		return Actor{}, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Coop-Internal-Secret", cfg.InternalSecret)

	resp, err := webVerifyClient.Do(req)
	if err != nil {
		return Actor{}, fmt.Errorf("request web app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Actor{}, fmt.Errorf("web app returned status %d", resp.StatusCode)
	}

	var result sessionVerifyResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Actor{}, fmt.Errorf("decode response: %w", err)
	}

	return Actor{UserID: result.UserID, DisplayName: result.Name}, nil
}

func RequireAnyIdentity(pool *db.Pool, cfg config.Config) func(http.HandlerFunc) http.HandlerFunc {
	cliMiddleware := RequireCliCredential(pool)
	browserMiddleware := RequireBrowserSession(cfg)

	return func(next http.HandlerFunc) http.HandlerFunc {
		cliNext := cliMiddleware(next)
		browserNext := browserMiddleware(next)

		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "" {
				cliNext(w, r)
				return
			}

			if _, err := r.Cookie(browserSessionCookieName); err == nil {
				browserNext(w, r)
				return
			}

			http.NotFound(w, r)
		}
	}
}
