package db_test

import (
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func newTestSession(t *testing.T, pool *db.Pool, id string) {
	t.Helper()

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop-"+id, "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := pool.CreateAgentSession(t.Context(), id, proj, "user-owner", "/repo", "/repo", "claude-code", time.Now()); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}
}

func TestAppendEventReservesIncreasingSeq(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	newTestSession(t, pool, "sess-a")

	first, err := pool.AppendEvent(t.Context(), "sess-a", []byte(`{"n":1}`))
	if err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
	second, err := pool.AppendEvent(t.Context(), "sess-a", []byte(`{"n":2}`))
	if err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}

	if first.Seq != 1 || second.Seq != 2 {
		t.Fatalf("got seqs %d, %d, want 1, 2", first.Seq, second.Seq)
	}
}

func TestEventsBeforeReturnsAscendingPageAndReversesQueryOrder(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	newTestSession(t, pool, "sess-a")

	for i := 0; i < 5; i++ {
		if _, err := pool.AppendEvent(t.Context(), "sess-a", []byte(`{}`)); err != nil {
			t.Fatalf("AppendEvent: %v", err)
		}
	}

	events, err := pool.EventsBefore(t.Context(), "sess-a", 4, 10)
	if err != nil {
		t.Fatalf("EventsBefore: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	for i, e := range events {
		if e.Seq != i+1 {
			t.Fatalf("event %d has seq %d, want ascending order starting at 1: %+v", i, e.Seq, events)
		}
	}
}

func TestEventsBeforeRespectsLimit(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	newTestSession(t, pool, "sess-a")

	for i := 0; i < 5; i++ {
		if _, err := pool.AppendEvent(t.Context(), "sess-a", []byte(`{}`)); err != nil {
			t.Fatalf("AppendEvent: %v", err)
		}
	}

	events, err := pool.EventsBefore(t.Context(), "sess-a", 6, 2)
	if err != nil {
		t.Fatalf("EventsBefore: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].Seq != 4 || events[1].Seq != 5 {
		t.Fatalf("got seqs %d, %d, want the 2 most recent below beforeSeq: 4, 5", events[0].Seq, events[1].Seq)
	}
}

func TestRecentEventsReturnsLastNInAscendingOrder(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	newTestSession(t, pool, "sess-a")

	for i := 0; i < 5; i++ {
		if _, err := pool.AppendEvent(t.Context(), "sess-a", []byte(`{}`)); err != nil {
			t.Fatalf("AppendEvent: %v", err)
		}
	}

	events, err := pool.RecentEvents(t.Context(), "sess-a", 3)
	if err != nil {
		t.Fatalf("RecentEvents: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	if events[0].Seq != 3 || events[1].Seq != 4 || events[2].Seq != 5 {
		t.Fatalf("unexpected seqs: %+v", events)
	}
}

func TestRecentEventsUnknownSessionReturnsEmpty(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	events, err := pool.RecentEvents(t.Context(), "sess-ghost", 10)
	if err != nil {
		t.Fatalf("RecentEvents: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}
