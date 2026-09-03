package stream

import (
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestTakeoverRegistryGetDefaultsToInactive(t *testing.T) {
	r := NewTakeoverRegistry(nil)

	got, err := r.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Active {
		t.Fatalf("got %+v, want inactive default", got)
	}
}

func TestTakeoverRegistrySetThenGet(t *testing.T) {
	r := NewTakeoverRegistry(nil)

	if err := r.Set(t.Context(), "sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := r.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Active || got.ByID != "u1" || got.By != "Alice" {
		t.Fatalf("got %+v, want active held by Alice", got)
	}
}

func TestTakeoverRegistryReleaseClearsState(t *testing.T) {
	r := NewTakeoverRegistry(nil)

	if err := r.Set(t.Context(), "sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := r.Set(t.Context(), "sess-a", TakeoverState{Active: false, ByID: "u1", By: "Alice"}); err != nil {
		t.Fatalf("Set (release): %v", err)
	}

	got, err := r.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Active {
		t.Fatalf("got %+v, want inactive after release", got)
	}
}

func TestTakeoverRegistryIsolatesSessions(t *testing.T) {
	r := NewTakeoverRegistry(nil)

	if err := r.Set(t.Context(), "sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := r.Get(t.Context(), "sess-b")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Active {
		t.Fatalf("got %+v, want session-b unaffected by session-a's takeover", got)
	}
}

func TestTakeoverRegistryConcurrentSetAndGet(t *testing.T) {
	r := NewTakeoverRegistry(nil)

	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			_ = r.Set(t.Context(), "sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"})
			_, _ = r.Get(t.Context(), "sess-a")
			done <- struct{}{}
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

func takeoverRegistryFixture(t *testing.T, sessionID string) *db.Pool {
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

func TestTakeoverRegistrySurvivesRestart(t *testing.T) {
	pool := takeoverRegistryFixture(t, "sess-a")

	r := NewTakeoverRegistry(pool)
	if err := r.Set(t.Context(), "sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// A fresh registry backed by the same pool simulates a relay restart: the in-memory cache is gone.
	restarted := NewTakeoverRegistry(pool)

	got, err := restarted.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get after restart: %v", err)
	}
	if !got.Active || got.ByID != "u1" || got.By != "Alice" {
		t.Fatalf("got %+v after restart, want takeover still held by Alice", got)
	}
}

func TestTakeoverRegistryReleaseSurvivesRestart(t *testing.T) {
	pool := takeoverRegistryFixture(t, "sess-a")

	r := NewTakeoverRegistry(pool)
	if err := r.Set(t.Context(), "sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := r.Set(t.Context(), "sess-a", TakeoverState{Active: false, ByID: "u1", By: "Alice"}); err != nil {
		t.Fatalf("Set (release): %v", err)
	}

	restarted := NewTakeoverRegistry(pool)

	got, err := restarted.Get(t.Context(), "sess-a")
	if err != nil {
		t.Fatalf("Get after restart: %v", err)
	}
	if got.Active {
		t.Fatalf("got %+v after restart, want inactive (release must persist too)", got)
	}
}
