package httpapi

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestHandleCLIExchangeReturnsTokenOnSuccess(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	webApp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/internal/users/resolve-github" {
			t.Errorf("got path %q, want /api/internal/users/resolve-github", r.URL.Path)
		}
		if got := r.Header.Get("X-Coop-Internal-Secret"); got != "test-secret" {
			t.Errorf("got secret %q, want test-secret", got)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"userId":      "user-123",
			"username":    "octocat",
			"displayName": "The Octocat",
			"avatarUrl":   "https://example.com/avatar.png",
		})
	}))
	defer webApp.Close()

	cfg := config.Config{InternalSecret: "test-secret", WebInternalURL: webApp.URL}

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cli/exchange", strings.NewReader(`{"github_access_token":"gh-token"}`))
	rec := httptest.NewRecorder()

	handleCLIExchange(cfg, pool)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var payload cliExchangeResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Username != "octocat" || payload.DisplayName != "The Octocat" || payload.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("unexpected payload: %+v", payload)
	}

	rawToken, err := hex.DecodeString(payload.Token)
	if err != nil {
		t.Fatalf("token is not hex: %v", err)
	}

	userID, displayName, err := pool.AuthenticateCliCredential(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("AuthenticateCliCredential: %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("got userID %q, want user-123", userID)
	}
	if displayName != "The Octocat" {
		t.Fatalf("got displayName %q, want The Octocat", displayName)
	}
}

func TestHandleCLIExchangeReturnsBadGatewayWhenWebAppFails(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	webApp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer webApp.Close()

	cfg := config.Config{InternalSecret: "test-secret", WebInternalURL: webApp.URL}

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cli/exchange", strings.NewReader(`{"github_access_token":"gh-token"}`))
	rec := httptest.NewRecorder()

	handleCLIExchange(cfg, pool)(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("got status %d, want 502: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleCLIExchangeValidation(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	cfg := config.Config{InternalSecret: "test-secret", WebInternalURL: "http://unused.invalid"}

	tests := []struct {
		name string
		body string
	}{
		{"missing token", `{}`},
		{"empty token", `{"github_access_token":""}`},
		{"malformed JSON", `{not json`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/cli/exchange", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handleCLIExchange(cfg, pool)(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
			}
		})
	}
}
