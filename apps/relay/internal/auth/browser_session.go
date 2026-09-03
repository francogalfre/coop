package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
)

const (
	browserSessionCookieName = "better-auth.session_token"
	webVerifyTimeout         = 10 * time.Second
	browserSessionCacheTTL   = 20 * time.Second
	browserSessionCacheCap   = 4096
)

var webVerifyClient = &http.Client{Timeout: webVerifyTimeout}

var browserSessionCache = newVerifyCache()

type verifyCacheEntry struct {
	actor     Actor
	expiresAt time.Time
}

type verifyCache struct {
	mu sync.RWMutex
	m  map[string]verifyCacheEntry
}

func newVerifyCache() *verifyCache {
	return &verifyCache{m: map[string]verifyCacheEntry{}}
}

func (c *verifyCache) get(cookieValue string) (Actor, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.m[cookieValue]
	if !ok || time.Now().After(entry.expiresAt) {
		return Actor{}, false
	}

	return entry.actor, true
}

// A cookie is one browser's whole visit, so a short cache turns a page's burst of relay calls (WS handshake, steer, presence) into one Postgres round trip instead of one per call.
func (c *verifyCache) set(cookieValue string, actor Actor) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.m) >= browserSessionCacheCap {
		c.m = map[string]verifyCacheEntry{}
	}

	c.m[cookieValue] = verifyCacheEntry{actor: actor, expiresAt: time.Now().Add(browserSessionCacheTTL)}
}

type sessionVerifyRequestBody struct {
	Cookie string `json:"cookie"`
}

type sessionVerifyResponseBody struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Image  string `json:"image"`
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
	cacheKey := cfg.WebInternalURL + "|" + cookieValue

	if actor, ok := browserSessionCache.get(cacheKey); ok {
		return actor, nil
	}

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

	actor := Actor{UserID: result.UserID, DisplayName: result.Name, AvatarURL: result.Image}
	browserSessionCache.set(cacheKey, actor)

	return actor, nil
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
