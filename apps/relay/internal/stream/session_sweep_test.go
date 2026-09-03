package stream

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestSweepStaleSessionsBroadcastsSessionEnd(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	ctx := t.Context()

	proj, err := pool.CreateProject(ctx, "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := pool.CreateAgentSession(ctx, "sess-stale", proj, "user-owner", "/repo", "/repo", "claude-code", time.Now().Add(-2*time.Hour)); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	store := New()
	events, unsubscribe := store.Subscribe("sess-stale")
	defer unsubscribe()

	ended, err := SweepStaleSessions(ctx, pool, store, time.Now().Add(-30*time.Minute))
	if err != nil {
		t.Fatalf("SweepStaleSessions: %v", err)
	}
	if len(ended) != 1 || ended[0] != "sess-stale" {
		t.Fatalf("got ended %v, want [sess-stale]", ended)
	}

	select {
	case e := <-events:
		var fields struct {
			Type      string `json:"type"`
			SessionID string `json:"session_id"`
			Seq       int    `json:"seq"`
		}
		if err := json.Unmarshal(e.Data, &fields); err != nil {
			t.Fatalf("unmarshal broadcast: %v", err)
		}
		if fields.Type != "session.end" || fields.SessionID != "sess-stale" || fields.Seq != e.Seq {
			t.Fatalf("got %+v (event seq %d), want session.end for sess-stale", fields, e.Seq)
		}
	default:
		t.Fatal("no event broadcast to subscriber")
	}

	sess, err := pool.GetAgentSession(ctx, "sess-stale")
	if err != nil {
		t.Fatalf("GetAgentSession: %v", err)
	}
	if sess.Status != db.SessionStatusEnded {
		t.Fatalf("got status %q, want ended", sess.Status)
	}
}
