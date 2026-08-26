package stream

import "testing"

func TestTakeoverRegistryGetDefaultsToInactive(t *testing.T) {
	r := NewTakeoverRegistry()

	got := r.Get("sess-a")
	if got.Active {
		t.Fatalf("got %+v, want inactive default", got)
	}
}

func TestTakeoverRegistrySetThenGet(t *testing.T) {
	r := NewTakeoverRegistry()

	r.Set("sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"})

	got := r.Get("sess-a")
	if !got.Active || got.ByID != "u1" || got.By != "Alice" {
		t.Fatalf("got %+v, want active held by Alice", got)
	}
}

func TestTakeoverRegistryReleaseClearsState(t *testing.T) {
	r := NewTakeoverRegistry()

	r.Set("sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"})
	r.Set("sess-a", TakeoverState{Active: false, ByID: "u1", By: "Alice"})

	got := r.Get("sess-a")
	if got.Active {
		t.Fatalf("got %+v, want inactive after release", got)
	}
}

func TestTakeoverRegistryIsolatesSessions(t *testing.T) {
	r := NewTakeoverRegistry()

	r.Set("sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"})

	got := r.Get("sess-b")
	if got.Active {
		t.Fatalf("got %+v, want session-b unaffected by session-a's takeover", got)
	}
}

func TestTakeoverRegistryConcurrentSetAndGet(t *testing.T) {
	r := NewTakeoverRegistry()

	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			r.Set("sess-a", TakeoverState{Active: true, ByID: "u1", By: "Alice"})
			r.Get("sess-a")
			done <- struct{}{}
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}
