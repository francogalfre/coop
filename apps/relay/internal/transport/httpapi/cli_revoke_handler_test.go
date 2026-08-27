package httpapi

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestHandleCLIRevokeInvalidatesCallersToken(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	rawToken, err := pool.CreateCliCredential(t.Context(), "user-123", "Alice")
	if err != nil {
		t.Fatalf("CreateCliCredential: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cli/revoke", nil)
	req.Header.Set("Authorization", "Bearer "+hex.EncodeToString(rawToken))
	req = withActor(req, "user-123")
	rec := httptest.NewRecorder()

	handleCLIRevoke(pool)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	if _, _, err := pool.AuthenticateCliCredential(t.Context(), rawToken); err == nil {
		t.Fatal("expected the revoked token to no longer authenticate")
	}
}

func TestHandleCLIRevokeRejectsMissingCredential(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/cli/revoke", nil)
	req = withActor(req, "user-123")
	rec := httptest.NewRecorder()

	handleCLIRevoke(pool)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}
