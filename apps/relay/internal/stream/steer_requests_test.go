package stream

import (
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestSteerRequestRegistryPutThenTake(t *testing.T) {
	r := NewSteerRequestRegistry(nil)

	req := PendingSteerRequest{RequestID: "req-1", Actor: auth.Actor{UserID: "u1", DisplayName: "Alice"}, Text: "hi"}
	if err := r.Put(t.Context(), "sess-a", req); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, ok, err := r.Take(t.Context(), "sess-a", "req-1")
	if err != nil {
		t.Fatalf("Take: %v", err)
	}
	if !ok || got != req {
		t.Fatalf("got %+v, %v, want %+v, true", got, ok, req)
	}

	if _, ok, err := r.Take(t.Context(), "sess-a", "req-1"); err != nil || ok {
		t.Fatalf("Take (second time): got ok=%v, err=%v, want one-shot miss", ok, err)
	}
}

func TestSteerRequestRegistryEvictsOldestOverCap(t *testing.T) {
	r := NewSteerRequestRegistry(nil)

	for i := 0; i < steerRequestCap+1; i++ {
		id := "req-" + string(rune('a'+i))
		if err := r.Put(t.Context(), "sess-a", PendingSteerRequest{RequestID: id, Text: "x"}); err != nil {
			t.Fatalf("Put(%s): %v", id, err)
		}
	}

	if _, ok, err := r.Take(t.Context(), "sess-a", "req-a"); err != nil || ok {
		t.Fatalf("Take(req-a): got ok=%v, err=%v, want it evicted as the oldest", ok, err)
	}

	if _, ok, err := r.Take(t.Context(), "sess-a", "req-"+string(rune('a'+steerRequestCap))); err != nil || !ok {
		t.Fatalf("Take(newest): got ok=%v, err=%v, want the newest request still pending", ok, err)
	}
}

func steerRequestRegistryFixture(t *testing.T, sessionID string) *db.Pool {
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

func TestSteerRequestRegistrySurvivesRestart(t *testing.T) {
	pool := steerRequestRegistryFixture(t, "sess-a")

	r := NewSteerRequestRegistry(pool)
	req := PendingSteerRequest{RequestID: "req-1", Actor: auth.Actor{UserID: "u1", DisplayName: "Alice"}, Text: "try the other branch"}
	if err := r.Put(t.Context(), "sess-a", req); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// A fresh registry backed by the same pool simulates a relay restart: the in-memory cache is gone.
	restarted := NewSteerRequestRegistry(pool)

	got, ok, err := restarted.Take(t.Context(), "sess-a", "req-1")
	if err != nil {
		t.Fatalf("Take after restart: %v", err)
	}
	if !ok {
		t.Fatal("got ok=false after restart, want the pending steer request still resolvable")
	}
	if got.Text != req.Text || got.Actor.UserID != req.Actor.UserID || got.Actor.DisplayName != req.Actor.DisplayName {
		t.Fatalf("got %+v after restart, want %+v", got, req)
	}
}

func TestSteerRequestRegistryEvictionSurvivesRestart(t *testing.T) {
	pool := steerRequestRegistryFixture(t, "sess-a")

	r := NewSteerRequestRegistry(pool)
	for i := 0; i < steerRequestCap+1; i++ {
		id := "req-" + string(rune('a'+i))
		req := PendingSteerRequest{RequestID: id, Actor: auth.Actor{UserID: "u1", DisplayName: "Someone"}, Text: "x"}
		if err := r.Put(t.Context(), "sess-a", req); err != nil {
			t.Fatalf("Put(%s): %v", id, err)
		}
	}

	restarted := NewSteerRequestRegistry(pool)

	if _, ok, err := restarted.Take(t.Context(), "sess-a", "req-a"); err != nil || ok {
		t.Fatalf("Take(req-a) after restart: got ok=%v, err=%v, want it stayed evicted", ok, err)
	}
}
