package stream

import (
	"encoding/json"
	"sync"
	"testing"
)

func mustAppend(t *testing.T, s *Store, sessionID string, event json.RawMessage) Event {
	t.Helper()

	recorded, err := s.Append(sessionID, event)
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	return recorded
}

func decodedSeq(t *testing.T, e Event) int {
	t.Helper()

	var fields struct {
		Seq int `json:"seq"`
	}
	if err := json.Unmarshal(e.Data, &fields); err != nil {
		t.Fatalf("failed to decode event data: %v", err)
	}

	return fields.Seq
}

func TestAppendAssignsIncreasingSeq(t *testing.T) {
	s := New()

	first := mustAppend(t, s, "sess-a", json.RawMessage(`{"n":1}`))
	second := mustAppend(t, s, "sess-a", json.RawMessage(`{"n":2}`))

	if first.Seq != 1 || second.Seq != 2 {
		t.Fatalf("got seqs %d, %d, want 1, 2", first.Seq, second.Seq)
	}
	if decodedSeq(t, first) != 1 || decodedSeq(t, second) != 2 {
		t.Fatalf("got encoded seqs %d, %d, want 1, 2", decodedSeq(t, first), decodedSeq(t, second))
	}
}

func TestAppendSeqIsPerSession(t *testing.T) {
	s := New()

	mustAppend(t, s, "sess-a", json.RawMessage(`{"n":1}`))
	first := mustAppend(t, s, "sess-b", json.RawMessage(`{"n":1}`))

	if first.Seq != 1 {
		t.Fatalf("got seq %d, want 1 for a fresh session", first.Seq)
	}
}

func TestAppendRewritesSeqFieldInStoredData(t *testing.T) {
	s := New()

	recorded := mustAppend(t, s, "sess-a", json.RawMessage(`{"n":1,"seq":0}`))

	if decodedSeq(t, recorded) != 1 {
		t.Fatalf("got encoded seq %d, want 1", decodedSeq(t, recorded))
	}
}

func TestAppendMalformedEventReturnsError(t *testing.T) {
	s := New()

	if _, err := s.Append("sess-a", json.RawMessage(`not json`)); err == nil {
		t.Fatal("expected error for malformed event")
	}
}

func TestSinceReturnsOrderedBacklog(t *testing.T) {
	s := New()

	mustAppend(t, s, "sess-a", json.RawMessage(`{"n":1}`))
	mustAppend(t, s, "sess-a", json.RawMessage(`{"n":2}`))
	mustAppend(t, s, "sess-a", json.RawMessage(`{"n":3}`))

	events := s.Since("sess-a", 0)
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	for i, e := range events {
		if e.Seq != i+1 {
			t.Fatalf("event %d has seq %d, want %d", i, e.Seq, i+1)
		}
		if decodedSeq(t, e) != i+1 {
			t.Fatalf("event %d has encoded seq %d, want %d", i, decodedSeq(t, e), i+1)
		}
	}
}

func TestSinceFiltersByAfterSeq(t *testing.T) {
	s := New()

	mustAppend(t, s, "sess-a", json.RawMessage(`{"n":1}`))
	mustAppend(t, s, "sess-a", json.RawMessage(`{"n":2}`))
	mustAppend(t, s, "sess-a", json.RawMessage(`{"n":3}`))

	events := s.Since("sess-a", 1)
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].Seq != 2 || events[1].Seq != 3 {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestSinceUnknownSessionReturnsEmpty(t *testing.T) {
	s := New()

	events := s.Since("sess-ghost", 0)
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestBufferEvictsOldestBeyondBound(t *testing.T) {
	s := New()

	for i := 0; i < bufferSize+10; i++ {
		mustAppend(t, s, "sess-a", json.RawMessage(`{}`))
	}

	events := s.Since("sess-a", 0)
	if len(events) != bufferSize {
		t.Fatalf("got %d events, want %d", len(events), bufferSize)
	}
	if events[0].Seq != 11 {
		t.Fatalf("oldest retained event has seq %d, want 11", events[0].Seq)
	}
	if events[len(events)-1].Seq != bufferSize+10 {
		t.Fatalf("newest retained event has seq %d, want %d", events[len(events)-1].Seq, bufferSize+10)
	}
}

func TestAppendWithSeqUsesGivenSeq(t *testing.T) {
	s := New()

	recorded, err := s.AppendWithSeq("sess-a", 42, json.RawMessage(`{"n":1}`))
	if err != nil {
		t.Fatalf("append with seq failed: %v", err)
	}

	if recorded.Seq != 42 {
		t.Fatalf("got seq %d, want 42", recorded.Seq)
	}
	if decodedSeq(t, recorded) != 42 {
		t.Fatalf("got encoded seq %d, want 42", decodedSeq(t, recorded))
	}
}

func TestAppendWithSeqKeepsStoreConsistentForSubsequentAppend(t *testing.T) {
	s := New()

	s.AppendWithSeq("sess-a", 5, json.RawMessage(`{"n":1}`))

	next := mustAppend(t, s, "sess-a", json.RawMessage(`{"n":2}`))
	if next.Seq != 6 {
		t.Fatalf("got seq %d, want 6 (nextSeq must track the DB-assigned seq)", next.Seq)
	}
}

func TestAppendWithSeqMalformedEventReturnsError(t *testing.T) {
	s := New()

	if _, err := s.AppendWithSeq("sess-a", 1, json.RawMessage(`not json`)); err == nil {
		t.Fatal("expected error for malformed event")
	}
}

func TestAppendWithSeqIsVisibleToSince(t *testing.T) {
	s := New()

	s.AppendWithSeq("sess-a", 1, json.RawMessage(`{"n":1}`))
	s.AppendWithSeq("sess-a", 2, json.RawMessage(`{"n":2}`))

	events := s.Since("sess-a", 0)
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].Seq != 1 || events[1].Seq != 2 {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestConcurrentAppendAndSince(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Append("sess-a", json.RawMessage(`{}`))
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Since("sess-a", 0)
		}()
	}

	wg.Wait()

	events := s.Since("sess-a", 0)
	if len(events) != 50 {
		t.Fatalf("got %d events, want 50", len(events))
	}
}
