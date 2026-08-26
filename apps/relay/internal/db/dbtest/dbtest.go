package dbtest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/francogalfre/coop/apps/relay/internal/db"
)

const defaultAdminURL = "postgres://postgres:password@localhost:5432/postgres"

func OpenScratch(t *testing.T) *db.Pool {
	t.Helper()

	ctx := context.Background()

	adminURL := os.Getenv("TEST_DATABASE_ADMIN_URL")
	if adminURL == "" {
		adminURL = defaultAdminURL
	}

	adminDB, err := sql.Open("pgx", adminURL)
	if err != nil {
		t.Skipf("dbtest: no Postgres available at %s: %v", adminURL, err)
	}

	if err := adminDB.PingContext(ctx); err != nil {
		adminDB.Close()
		t.Skipf("dbtest: no Postgres available at %s: %v", adminURL, err)
	}

	name := scratchName(t)

	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %q`, name)); err != nil {
		adminDB.Close()
		t.Fatalf("dbtest: create database %s: %v", name, err)
	}

	t.Cleanup(func() {
		defer adminDB.Close()
		if _, err := adminDB.ExecContext(context.Background(), fmt.Sprintf(`DROP DATABASE %q WITH (FORCE)`, name)); err != nil {
			t.Logf("dbtest: drop database %s: %v", name, err)
		}
	})

	pool, err := db.Open(ctx, replaceDatabaseName(adminURL, name))
	if err != nil {
		t.Fatalf("dbtest: open scratch pool: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	if err := pool.Migrate(ctx); err != nil {
		t.Fatalf("dbtest: migrate scratch database: %v", err)
	}

	return pool
}

func scratchName(t *testing.T) string {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		t.Fatalf("dbtest: generate scratch db name: %v", err)
	}

	return "coop_test_" + hex.EncodeToString(raw)
}

func replaceDatabaseName(adminURL, name string) string {
	for i := len(adminURL) - 1; i >= 0; i-- {
		if adminURL[i] == '/' {
			return adminURL[:i+1] + name
		}
	}
	return adminURL + "/" + name
}
