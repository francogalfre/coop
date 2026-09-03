package auth_test

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func newBrowserVerifyServer(t *testing.T, wantSecret string, cookieToUserID map[string]string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Coop-Internal-Secret") != wantSecret {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var body struct {
			Cookie string `json:"cookie"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userID, ok := cookieToUserID[body.Cookie]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).
			Encode(map[string]string{"userId": userID, "name": "Display " + userID, "image": "https://avatars.example/" + userID})
	}))
}

func TestRequireBrowserSessionAcceptsValidCookie(t *testing.T) {
	srv := newBrowserVerifyServer(t, "s3cret", map[string]string{"good-cookie": "user-42"})
	defer srv.Close()

	cfg := config.Config{WebInternalURL: srv.URL, InternalSecret: "s3cret"}

	var gotActor auth.Actor

	handler := auth.RequireBrowserSession(cfg)(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			t.Fatal("FromContext: actor not set")
		}
		gotActor = actor
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "good-cookie"})
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}

	if gotActor.UserID != "user-42" {
		t.Fatalf("got actor.UserID %q, want %q", gotActor.UserID, "user-42")
	}
	if gotActor.DisplayName != "Display user-42" {
		t.Fatalf("got actor.DisplayName %q, want %q", gotActor.DisplayName, "Display user-42")
	}
	if gotActor.AvatarURL != "https://avatars.example/user-42" {
		t.Fatalf("got actor.AvatarURL %q, want %q", gotActor.AvatarURL, "https://avatars.example/user-42")
	}
}

func TestRequireBrowserSessionCachesSuccessfulVerification(t *testing.T) {
	var calls atomic.Int32
	cookie := "cached-cookie-" + t.Name()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"userId": "user-cached", "name": "Cached User"})
	}))
	defer srv.Close()

	cfg := config.Config{WebInternalURL: srv.URL, InternalSecret: "s3cret"}
	handler := auth.RequireBrowserSession(cfg)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: cookie})
		rec := httptest.NewRecorder()
		handler(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want 200", rec.Code)
		}
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("web app was called %d times, want 1 (later calls should hit the cache)", got)
	}
}

func TestRequireBrowserSessionRejectsMissingCookie(t *testing.T) {
	srv := newBrowserVerifyServer(t, "s3cret", nil)
	defer srv.Close()

	cfg := config.Config{WebInternalURL: srv.URL, InternalSecret: "s3cret"}

	handler := auth.RequireBrowserSession(cfg)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestRequireBrowserSessionRejectsUnresolvedCookie(t *testing.T) {
	srv := newBrowserVerifyServer(t, "s3cret", map[string]string{})
	defer srv.Close()

	cfg := config.Config{WebInternalURL: srv.URL, InternalSecret: "s3cret"}

	handler := auth.RequireBrowserSession(cfg)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "unknown-cookie"})
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestRequireAnyIdentityAcceptsCliCredential(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	rawToken, err := pool.CreateCliCredential(t.Context(), "user-cli", "CLI User")
	if err != nil {
		t.Fatalf("CreateCliCredential: %v", err)
	}

	cfg := config.Config{WebInternalURL: "http://unused.invalid", InternalSecret: "s3cret"}

	var gotActor auth.Actor

	handler := auth.RequireAnyIdentity(pool, cfg)(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			t.Fatal("FromContext: actor not set")
		}
		gotActor = actor
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+hex.EncodeToString(rawToken))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	if gotActor.UserID != "user-cli" {
		t.Fatalf("got actor.UserID %q, want %q", gotActor.UserID, "user-cli")
	}
}

func TestRequireAnyIdentityRejectsNoCredentials(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	cfg := config.Config{WebInternalURL: "http://unused.invalid", InternalSecret: "s3cret"}

	handler := auth.RequireAnyIdentity(pool, cfg)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}
