package db

import (
	"testing"
	"time"
)

func TestStaleSessionSweepFlipsOnlyIdleLiveSessions(t *testing.T) {
	pool := openScratchPool(t)
	ctx := t.Context()

	proj, err := pool.CreateProject(ctx, "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	now := time.Now()
	idleSince := now.Add(-30 * time.Minute)

	mkSession := func(id string, startedAt time.Time) {
		if _, err := pool.CreateAgentSession(ctx, id, proj, "user-owner", "/repo", "/repo", "claude-code", startedAt); err != nil {
			t.Fatalf("CreateAgentSession(%s): %v", id, err)
		}
	}

	mkEvent := func(sessionID string, seq int, createdAt time.Time) {
		if _, err := pool.client.Event.Create().
			SetSessionID(sessionID).
			SetSeq(seq).
			SetData([]byte(`{"type":"tool.started"}`)).
			SetCreatedAt(createdAt).
			Save(ctx); err != nil {
			t.Fatalf("create event on %s: %v", sessionID, err)
		}
	}

	mkSession("stale-no-events", now.Add(-2*time.Hour))
	mkSession("stale-old-events", now.Add(-3*time.Hour))
	mkEvent("stale-old-events", 1, now.Add(-90*time.Minute))
	mkSession("fresh-events", now.Add(-3*time.Hour))
	mkEvent("fresh-events", 1, now.Add(-5*time.Minute))
	mkSession("fresh-start", now.Add(-2*time.Minute))

	mkSession("already-ended", now.Add(-4*time.Hour))
	if err := pool.EndAgentSession(ctx, "already-ended", now.Add(-3*time.Hour)); err != nil {
		t.Fatalf("EndAgentSession(already-ended): %v", err)
	}

	ids, err := pool.StaleSessionIDs(ctx, idleSince)
	if err != nil {
		t.Fatalf("StaleSessionIDs: %v", err)
	}

	got := map[string]bool{}
	for _, id := range ids {
		got[id] = true
	}
	if len(ids) != 2 || !got["stale-no-events"] || !got["stale-old-events"] {
		t.Fatalf("got stale ids %v, want [stale-no-events stale-old-events]", ids)
	}

	for _, id := range ids {
		ended, err := pool.EndSessionIfStale(ctx, id, idleSince)
		if err != nil || !ended {
			t.Fatalf("EndSessionIfStale(%s): ended=%v err=%v", id, ended, err)
		}
	}

	assertStatus := func(id, want string) {
		sess, err := pool.GetAgentSession(ctx, id)
		if err != nil {
			t.Fatalf("GetAgentSession(%s): %v", id, err)
		}
		if sess.Status != want {
			t.Fatalf("%s: got status %q, want %q", id, sess.Status, want)
		}
		if want == SessionStatusEnded && sess.EndedAt == nil {
			t.Fatalf("%s: status ended but ended_at is nil", id)
		}
	}

	assertStatus("stale-no-events", SessionStatusEnded)
	assertStatus("stale-old-events", SessionStatusEnded)
	assertStatus("fresh-events", SessionStatusLive)
	assertStatus("fresh-start", SessionStatusLive)
}

func TestEndSessionIfStaleSkipsSessionThatGotAnEventSinceListing(t *testing.T) {
	pool := openScratchPool(t)
	ctx := t.Context()

	proj, err := pool.CreateProject(ctx, "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	now := time.Now()
	idleSince := now.Add(-30 * time.Minute)

	if _, err := pool.CreateAgentSession(ctx, "raced", proj, "user-owner", "/repo", "/repo", "claude-code", now.Add(-2*time.Hour)); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	ids, err := pool.StaleSessionIDs(ctx, idleSince)
	if err != nil || len(ids) != 1 {
		t.Fatalf("StaleSessionIDs: ids=%v err=%v", ids, err)
	}

	if _, err := pool.client.Event.Create().
		SetSessionID("raced").
		SetSeq(1).
		SetData([]byte(`{"type":"tool.started"}`)).
		SetCreatedAt(now).
		Save(ctx); err != nil {
		t.Fatalf("create event: %v", err)
	}

	ended, err := pool.EndSessionIfStale(ctx, "raced", idleSince)
	if err != nil {
		t.Fatalf("EndSessionIfStale: %v", err)
	}
	if ended {
		t.Fatal("session that received an event since listing was still ended")
	}

	sess, err := pool.GetAgentSession(ctx, "raced")
	if err != nil {
		t.Fatalf("GetAgentSession: %v", err)
	}
	if sess.Status != SessionStatusLive {
		t.Fatalf("got status %q, want live", sess.Status)
	}
}
