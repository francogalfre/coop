package stream

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestSubscribeReceivesLiveAppends(t *testing.T) {
	s := New()

	ch, unsubscribe := s.Subscribe("sess-a")
	defer unsubscribe()

	s.Append("sess-a", json.RawMessage(`{"n":1}`))

	select {
	case event := <-ch:
		if event.Seq != 1 {
			t.Fatalf("got seq %d, want 1", event.Seq)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for live event")
	}
}

func TestSubscribeDoesNotReceiveOtherSessions(t *testing.T) {
	s := New()

	ch, unsubscribe := s.Subscribe("sess-a")
	defer unsubscribe()

	s.Append("sess-b", json.RawMessage(`{"n":1}`))

	select {
	case event := <-ch:
		t.Fatalf("got unexpected event for other session: %+v", event)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestUnsubscribeClosesChannel(t *testing.T) {
	s := New()

	ch, unsubscribe := s.Subscribe("sess-a")
	unsubscribe()

	_, open := <-ch
	if open {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}

func TestConcurrentSubscribeAndAppend(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch, unsubscribe := s.Subscribe("sess-a")
			defer unsubscribe()

			select {
			case <-ch:
			case <-time.After(time.Second):
			}
		}()
	}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Append("sess-a", json.RawMessage(`{}`))
		}()
	}

	wg.Wait()
}
