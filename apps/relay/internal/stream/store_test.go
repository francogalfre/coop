package stream

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestAppendAssignsIncreasingSeq(t *testing.T) {
	s := New()

	first := s.Append("sess-a", json.RawMessage(`{"n":1}`))
	second := s.Append("sess-a", json.RawMessage(`{"n":2}`))

	if first != 1 || second != 2 {
		t.Fatalf("got seqs %d, %d, want 1, 2", first, second)
	}
}

func TestAppendSeqIsPerSession(t *testing.T) {
	s := New()

	s.Append("sess-a", json.RawMessage(`{"n":1}`))
	first := s.Append("sess-b", json.RawMessage(`{"n":1}`))

	if first != 1 {
		t.Fatalf("got seq %d, want 1 for a fresh session", first)
	}
}

func TestSinceReturnsOrderedBacklog(t *testing.T) {
	s := New()

	s.Append("sess-a", json.RawMessage(`{"n":1}`))
	s.Append("sess-a", json.RawMessage(`{"n":2}`))
	s.Append("sess-a", json.RawMessage(`{"n":3}`))

	events := s.Since("sess-a", 0)
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	for i, e := range events {
		if e.Seq != i+1 {
			t.Fatalf("event %d has seq %d, want %d", i, e.Seq, i+1)
		}
	}
}

func TestSinceFiltersByAfterSeq(t *testing.T) {
	s := New()

	s.Append("sess-a", json.RawMessage(`{"n":1}`))
	s.Append("sess-a", json.RawMessage(`{"n":2}`))
	s.Append("sess-a", json.RawMessage(`{"n":3}`))

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
		s.Append("sess-a", json.RawMessage(`{}`))
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
