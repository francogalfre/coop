package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithCORSAllowsConfiguredOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := withCORS(next, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want http://localhost:3000", got)
	}
}

func TestWithCORSRejectsUnlistedOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := withCORS(next, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty for unlisted origin", got)
	}
}

func TestWithCORSNeverAllowsWildcard(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := withCORS(next, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "*" {
		t.Fatal("Access-Control-Allow-Origin must never be * once credentials are involved")
	}
}

func TestWithCORSAnswersPreflightWithNoContent(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := withCORS(next, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodOptions, "/v1/events", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if called {
		t.Fatal("preflight must not reach the wrapped handler")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("preflight response missing Access-Control-Allow-Headers")
	}
}
