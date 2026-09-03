package db_test

import (
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func takeoverFixture(t *testing.T, sessionID string) *db.Pool {
	t.Helper()

	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := pool.CreateAgentSession(t.Context(), sessionID, proj, "user-owner", "/repo", "/repo", "claude-code", time.Now()); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	return pool
}

func TestSetAndGetTakeoverRoundTrips(t *testing.T) {
	pool := takeoverFixture(t, "sess-a")

	if err := pool.SetTakeover(t.Context(), "sess-a", "user-alice", "Alice"); err != nil {
		t.Fatalf("SetTakeover: %v", err)
	}

	got, err := pool.GetTakeover(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("GetTakeover: %v", err)
	}
	if got.ActorID != "user-alice" || got.ActorDisplayName != "Alice" {
		t.Fatalf("got takeover %+v, want held by user-alice/Alice", got)
	}
}

func TestSetTakeoverReplacesExistingHolder(t *testing.T) {
	pool := takeoverFixture(t, "sess-a")

	if err := pool.SetTakeover(t.Context(), "sess-a", "user-alice", "Alice"); err != nil {
		t.Fatalf("SetTakeover(alice): %v", err)
	}
	if err := pool.SetTakeover(t.Context(), "sess-a", "user-bob", "Bob"); err != nil {
		t.Fatalf("SetTakeover(bob): %v", err)
	}

	got, err := pool.GetTakeover(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("GetTakeover: %v", err)
	}
	if got.ActorID != "user-bob" || got.ActorDisplayName != "Bob" {
		t.Fatalf("got takeover %+v, want held by user-bob/Bob", got)
	}
}

func TestClearTakeoverDeletesRow(t *testing.T) {
	pool := takeoverFixture(t, "sess-a")

	if err := pool.SetTakeover(t.Context(), "sess-a", "user-alice", "Alice"); err != nil {
		t.Fatalf("SetTakeover: %v", err)
	}
	if err := pool.ClearTakeover(t.Context(), "sess-a"); err != nil {
		t.Fatalf("ClearTakeover: %v", err)
	}

	if _, err := pool.GetTakeover(t.Context(), "sess-a"); !db.IsNotFound(err) {
		t.Fatalf("got err %v, want a not-found error after clear", err)
	}
}

func TestClearTakeoverIsIdempotent(t *testing.T) {
	pool := takeoverFixture(t, "sess-a")

	if err := pool.ClearTakeover(t.Context(), "sess-a"); err != nil {
		t.Fatalf("ClearTakeover (no row): %v", err)
	}
}

func TestGetTakeoverNotFoundWithoutRow(t *testing.T) {
	pool := takeoverFixture(t, "sess-a")

	if _, err := pool.GetTakeover(t.Context(), "sess-a"); !db.IsNotFound(err) {
		t.Fatalf("got err %v, want a not-found error", err)
	}
}
