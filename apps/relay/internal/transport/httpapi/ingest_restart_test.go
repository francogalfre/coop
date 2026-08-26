package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func TestHandleIngestPersistsSessionStartData(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	store := stream.New()
	ingest := handleIngest(pool, registry, store)

	now := time.Now().UTC()
	startBody := `{"v":1,"session_id":"sess-identity","seq":0,"ts":"` + now.Format(time.RFC3339) + `","type":"session.start","cwd":"/repo","repo":"coop/coop","owner":{"id":"alice","display_name":"Alice"},"harness":"claude-code"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(startBody))
	req.Header.Set(coopProjectHeader, "coop")
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	ingest(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	events, err := pool.EventsSince(t.Context(), "sess-identity", 0, 10)
	if err != nil {
		t.Fatalf("EventsSince: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected the replayed history to contain session.start, got %d events", len(events))
	}

	var fields map[string]any
	if err := json.Unmarshal(events[0].Data, &fields); err != nil {
		t.Fatalf("failed to decode persisted event: %v", err)
	}

	if fields["type"] != "session.start" || fields["harness"] != "claude-code" || fields["cwd"] != "/repo" {
		t.Fatalf("replayed session.start missing identity fields: %+v", fields)
	}
	if events[0].Seq != 1 {
		t.Fatalf("expected persisted session.start to reserve DB seq 1, got %d", events[0].Seq)
	}
}

func TestHandleIngestSurvivesRestartMidSession(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	store := stream.New()
	ingest := handleIngest(pool, registry, store)

	now := time.Now().UTC()
	startBody := `{"v":1,"session_id":"sess-restart","seq":0,"ts":"` + now.Format(time.RFC3339) + `","type":"session.start","cwd":"/repo","owner":{"id":"alice","display_name":"Alice"},"harness":"claude-code"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(startBody))
	req.Header.Set(coopProjectHeader, "coop")
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	ingest(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	// Fresh registry/store/handler simulates a restart: the persist cache is empty even though Postgres has the session.
	freshRegistry := presence.New()
	freshStore := stream.New()
	freshIngest := handleIngest(pool, freshRegistry, freshStore)

	// agent.text needs no presence-registry membership, unlike file.touched, so it isolates the persistence path.
	textBody := `{"v":1,"session_id":"sess-restart","seq":1,"ts":"` + now.Add(time.Second).Format(time.RFC3339) + `","type":"agent.text"}`

	req2 := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(textBody))
	req2.Header.Set(coopProjectHeader, "coop")
	rec2 := httptest.NewRecorder()

	freshIngest(rec2, req2)
	if rec2.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec2.Code, rec2.Body.String())
	}

	events, err := pool.EventsSince(t.Context(), "sess-restart", 0, 10)
	if err != nil {
		t.Fatalf("EventsSince: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected the post-restart event to still be persisted, got %d persisted events", len(events))
	}
}

func TestHandleIngestSeqAgreesBetweenPostgresAndStore(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	registry := presence.New()
	store := stream.New()
	ingest := handleIngest(pool, registry, store)

	now := time.Now().UTC()
	startBody := `{"v":1,"session_id":"sess-seq","seq":0,"ts":"` + now.Format(time.RFC3339) + `","type":"session.start","cwd":"/repo","owner":{"id":"alice","display_name":"Alice"},"harness":"claude-code"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(startBody))
	req.Header.Set(coopProjectHeader, "coop")
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()
	ingest(rec, req)

	touchedBody := `{"v":1,"session_id":"sess-seq","seq":99,"ts":"` + now.Add(time.Second).Format(time.RFC3339) + `","type":"file.touched","path":"src/foo.ts","mode":"write"}`

	req2 := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(touchedBody))
	req2.Header.Set(coopProjectHeader, "coop")
	rec2 := httptest.NewRecorder()
	ingest(rec2, req2)

	if rec2.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec2.Code, rec2.Body.String())
	}

	dbEvents, err := pool.EventsSince(t.Context(), "sess-seq", 0, 10)
	if err != nil {
		t.Fatalf("EventsSince: %v", err)
	}
	if len(dbEvents) != 2 {
		t.Fatalf("expected 2 persisted events, got %d", len(dbEvents))
	}

	memEvents := store.Since("sess-seq", 0)
	if len(memEvents) != 2 {
		t.Fatalf("expected 2 in-memory events, got %d", len(memEvents))
	}

	for i, dbEvent := range dbEvents {
		if memEvents[i].Seq != dbEvent.Seq {
			t.Fatalf("event %d: postgres seq %d disagrees with in-memory store seq %d", i, dbEvent.Seq, memEvents[i].Seq)
		}

		var memFields map[string]any
		if err := json.Unmarshal(memEvents[i].Data, &memFields); err != nil {
			t.Fatalf("failed to decode in-memory event: %v", err)
		}
		if int(memFields["seq"].(float64)) != dbEvent.Seq {
			t.Fatalf("event %d: seq embedded in the WS-broadcast JSON (%v) disagrees with postgres seq %d", i, memFields["seq"], dbEvent.Seq)
		}
	}

	// Postgres is the sole seq authority; the client-supplied seq (99) must never leak through.
	if memEvents[1].Seq == 99 {
		t.Fatal("client-supplied seq leaked through instead of the DB-assigned seq")
	}
}
