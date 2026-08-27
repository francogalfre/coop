package db

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testAdminURL = "postgres://postgres:password@localhost:5432/postgres"

// Mirrors dbtest.OpenScratch without importing the dbtest package, which itself imports db and would create an import cycle for an internal (package db) test.
func openScratchPool(t *testing.T) *Pool {
	t.Helper()

	ctx := context.Background()

	adminDB, err := sql.Open("pgx", testAdminURL)
	if err != nil {
		t.Skipf("no Postgres available at %s: %v", testAdminURL, err)
	}
	if err := adminDB.PingContext(ctx); err != nil {
		adminDB.Close()
		t.Skipf("no Postgres available at %s: %v", testAdminURL, err)
	}

	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		t.Fatalf("generate scratch db name: %v", err)
	}
	name := "coop_test_" + hex.EncodeToString(raw)

	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %q`, name)); err != nil {
		adminDB.Close()
		t.Fatalf("create database %s: %v", name, err)
	}
	t.Cleanup(func() {
		defer adminDB.Close()
		if _, err := adminDB.ExecContext(context.Background(), fmt.Sprintf(`DROP DATABASE %q WITH (FORCE)`, name)); err != nil {
			t.Logf("drop database %s: %v", name, err)
		}
	})

	pool, err := Open(ctx, scratchURL(testAdminURL, name))
	if err != nil {
		t.Fatalf("open scratch pool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	if err := pool.Migrate(ctx); err != nil {
		t.Fatalf("migrate scratch database: %v", err)
	}

	return pool
}

func scratchURL(adminURL, name string) string {
	for i := len(adminURL) - 1; i >= 0; i-- {
		if adminURL[i] == '/' {
			return adminURL[:i+1] + name
		}
	}
	return adminURL + "/" + name
}

func TestCreateCliCredentialSetsExpiry(t *testing.T) {
	pool := openScratchPool(t)

	if _, err := pool.CreateCliCredential(t.Context(), "user-123", "Alice"); err != nil {
		t.Fatalf("CreateCliCredential: %v", err)
	}

	cred, err := pool.client.CliCredential.Query().Only(t.Context())
	if err != nil {
		t.Fatalf("query credential: %v", err)
	}

	if cred.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set, got nil (never expires)")
	}
	if cred.ExpiresAt.Before(time.Now().Add(89*24*time.Hour)) || cred.ExpiresAt.After(time.Now().Add(91*24*time.Hour)) {
		t.Fatalf("expected ExpiresAt around 90 days out, got %v", cred.ExpiresAt)
	}
}

func TestAuthenticateCliCredentialRejectsExpiredToken(t *testing.T) {
	pool := openScratchPool(t)

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		t.Fatalf("generate token: %v", err)
	}
	sum := sha256.Sum256(raw)

	if _, err := pool.client.CliCredential.Create().
		SetUserID("user-123").
		SetDisplayName("Alice").
		SetTokenHash(sum[:]).
		SetExpiresAt(time.Now().Add(-time.Hour)).
		Save(t.Context()); err != nil {
		t.Fatalf("create expired credential: %v", err)
	}

	if _, _, err := pool.AuthenticateCliCredential(t.Context(), raw); err == nil {
		t.Fatal("expected authentication with an expired credential to fail")
	}
}
