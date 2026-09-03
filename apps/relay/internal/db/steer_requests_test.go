package db_test

import (
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func steerRequestFixture(t *testing.T, sessionID string) *db.Pool {
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

func TestCreateAndGetSteerRequestRoundTrips(t *testing.T) {
	pool := steerRequestFixture(t, "sess-a")

	if err := pool.CreateSteerRequest(t.Context(), "req-1", "sess-a", "user-alice", "Alice", "https://example.com/a.png", "try the other branch"); err != nil {
		t.Fatalf("CreateSteerRequest: %v", err)
	}

	got, err := pool.GetSteerRequest(t.Context(), "sess-a", "req-1")
	if err != nil {
		t.Fatalf("GetSteerRequest: %v", err)
	}

	if got.ActorID != "user-alice" || got.ActorDisplayName != "Alice" || got.ActorAvatarURL != "https://example.com/a.png" || got.Text != "try the other branch" {
		t.Fatalf("got steer request %+v, want matching stored fields", got)
	}
}

func TestGetSteerRequestNotFoundForWrongSession(t *testing.T) {
	pool := steerRequestFixture(t, "sess-a")

	proj, err := pool.GetProjectBySlug(t.Context(), "coop")
	if err != nil {
		t.Fatalf("GetProjectBySlug: %v", err)
	}
	if _, err := pool.CreateAgentSession(t.Context(), "sess-b", proj, "user-owner", "/repo", "/repo", "claude-code", time.Now()); err != nil {
		t.Fatalf("CreateAgentSession(sess-b): %v", err)
	}

	if err := pool.CreateSteerRequest(t.Context(), "req-1", "sess-a", "user-alice", "Alice", "", "text"); err != nil {
		t.Fatalf("CreateSteerRequest: %v", err)
	}

	if _, err := pool.GetSteerRequest(t.Context(), "sess-b", "req-1"); !db.IsNotFound(err) {
		t.Fatalf("got err %v, want a not-found error for a request scoped to a different session", err)
	}
}

func TestDeleteSteerRequestIsIdempotent(t *testing.T) {
	pool := steerRequestFixture(t, "sess-a")

	if err := pool.CreateSteerRequest(t.Context(), "req-1", "sess-a", "user-alice", "Alice", "", "text"); err != nil {
		t.Fatalf("CreateSteerRequest: %v", err)
	}

	if err := pool.DeleteSteerRequest(t.Context(), "req-1"); err != nil {
		t.Fatalf("DeleteSteerRequest: %v", err)
	}

	if err := pool.DeleteSteerRequest(t.Context(), "req-1"); err != nil {
		t.Fatalf("DeleteSteerRequest (already gone): %v", err)
	}

	if _, err := pool.GetSteerRequest(t.Context(), "sess-a", "req-1"); !db.IsNotFound(err) {
		t.Fatalf("got err %v, want a not-found error after delete", err)
	}
}

func TestEvictOldestSteerRequestsDropsOverflow(t *testing.T) {
	pool := steerRequestFixture(t, "sess-a")

	for i := 0; i < 5; i++ {
		id := "req-" + string(rune('a'+i))
		if err := pool.CreateSteerRequest(t.Context(), id, "sess-a", "user-alice", "Alice", "", "text"); err != nil {
			t.Fatalf("CreateSteerRequest(%s): %v", id, err)
		}
	}

	dropped, err := pool.EvictOldestSteerRequests(t.Context(), "sess-a", 2)
	if err != nil {
		t.Fatalf("EvictOldestSteerRequests: %v", err)
	}
	if len(dropped) != 3 {
		t.Fatalf("got %d dropped ids, want 3", len(dropped))
	}

	if _, err := pool.GetSteerRequest(t.Context(), "sess-a", "req-a"); !db.IsNotFound(err) {
		t.Fatalf("got err %v, want req-a (oldest) evicted", err)
	}
	if _, err := pool.GetSteerRequest(t.Context(), "sess-a", "req-e"); err != nil {
		t.Fatalf("GetSteerRequest(req-e): %v, want the newest request kept", err)
	}
}

func TestEvictOldestSteerRequestsNoOpUnderCap(t *testing.T) {
	pool := steerRequestFixture(t, "sess-a")

	if err := pool.CreateSteerRequest(t.Context(), "req-1", "sess-a", "user-alice", "Alice", "", "text"); err != nil {
		t.Fatalf("CreateSteerRequest: %v", err)
	}

	dropped, err := pool.EvictOldestSteerRequests(t.Context(), "sess-a", 16)
	if err != nil {
		t.Fatalf("EvictOldestSteerRequests: %v", err)
	}
	if len(dropped) != 0 {
		t.Fatalf("got %d dropped ids, want 0 under cap", len(dropped))
	}
}
