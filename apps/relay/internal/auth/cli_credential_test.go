package auth_test

import (
	"context"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestRequireCliCredentialAcceptsValidToken(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	rawToken, err := pool.CreateCliCredential(context.Background(), "user-123", "Alice")
	if err != nil {
		t.Fatalf("CreateCliCredential: %v", err)
	}

	var gotActor auth.Actor

	handler := auth.RequireCliCredential(pool)(func(w http.ResponseWriter, r *http.Request) {
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
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}

	if gotActor.UserID != "user-123" {
		t.Fatalf("got actor.UserID %q, want %q", gotActor.UserID, "user-123")
	}
	if gotActor.DisplayName != "Alice" {
		t.Fatalf("got actor.DisplayName %q, want %q", gotActor.DisplayName, "Alice")
	}
}

func TestRequireCliCredentialRejectsMissingHeader(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	handler := auth.RequireCliCredential(pool)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRequireCliCredentialRejectsMalformedHex(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	handler := auth.RequireCliCredential(pool)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-hex")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRequireCliCredentialRejectsUnknownToken(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	handler := auth.RequireCliCredential(pool)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	unknown := make([]byte, 32)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+hex.EncodeToString(unknown))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}
