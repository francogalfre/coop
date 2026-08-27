package db_test

import (
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestRevokeCliCredentialInvalidatesToken(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	rawToken, err := pool.CreateCliCredential(t.Context(), "user-123", "Alice")
	if err != nil {
		t.Fatalf("CreateCliCredential: %v", err)
	}

	if err := pool.RevokeCliCredential(t.Context(), "user-123", rawToken); err != nil {
		t.Fatalf("RevokeCliCredential: %v", err)
	}

	if _, _, err := pool.AuthenticateCliCredential(t.Context(), rawToken); err == nil {
		t.Fatal("expected authentication with a revoked token to fail")
	}
}

func TestRevokeCliCredentialDoesNotAffectOtherUsersCredential(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	rawToken, err := pool.CreateCliCredential(t.Context(), "user-victim", "Alice")
	if err != nil {
		t.Fatalf("CreateCliCredential: %v", err)
	}

	if err := pool.RevokeCliCredential(t.Context(), "user-attacker", rawToken); err != nil {
		t.Fatalf("RevokeCliCredential: %v", err)
	}

	if _, _, err := pool.AuthenticateCliCredential(t.Context(), rawToken); err != nil {
		t.Fatalf("expected victim's token to remain valid after another user's revoke attempt, got: %v", err)
	}
}

func TestRevokeCliCredentialIsIdempotent(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	unknown := make([]byte, 32)

	if err := pool.RevokeCliCredential(t.Context(), "user-nobody", unknown); err != nil {
		t.Fatalf("expected revoking an unknown token to succeed silently, got: %v", err)
	}
}
