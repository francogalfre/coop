package stream

import (
	"testing"
	"time"
)

func TestPtyHubBroadcastReachesViewer(t *testing.T) {
	h := NewPtyHub()

	viewer, unsub := h.Subscribe("sess-a")
	defer unsub()

	h.Broadcast("sess-a", []byte("output"))

	select {
	case msg := <-viewer:
		if string(msg) != "output" {
			t.Fatalf("got %q, want output", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broadcast")
	}
}

func TestPtyHubBroadcastFansOutToMultipleViewers(t *testing.T) {
	h := NewPtyHub()

	a, unsubA := h.Subscribe("sess-a")
	defer unsubA()
	b, unsubB := h.Subscribe("sess-a")
	defer unsubB()

	h.Broadcast("sess-a", []byte("output"))

	for _, ch := range []chan []byte{a, b} {
		select {
		case msg := <-ch:
			if string(msg) != "output" {
				t.Fatalf("got %q, want output", msg)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for broadcast")
		}
	}
}

func TestPtyHubBroadcastDoesNotReachOtherSessions(t *testing.T) {
	h := NewPtyHub()

	ch, unsub := h.Subscribe("sess-a")
	defer unsub()

	h.Broadcast("sess-b", []byte("output"))

	select {
	case msg := <-ch:
		t.Fatalf("got unexpected message for other session: %q", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestPtyHubSubscribeUnsubscribeClosesChannel(t *testing.T) {
	h := NewPtyHub()

	ch, unsub := h.Subscribe("sess-a")
	unsub()

	_, open := <-ch
	if open {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}

func TestPtyHubRouteInputDeliversToSource(t *testing.T) {
	h := NewPtyHub()

	deliver, unregister := h.SetSource("sess-a")
	defer unregister()

	if ok := h.RouteInput("sess-a", []byte("input")); !ok {
		t.Fatal("expected RouteInput to report delivery")
	}

	select {
	case msg := <-deliver:
		if string(msg) != "input" {
			t.Fatalf("got %q, want input", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for routed input")
	}
}

func TestPtyHubRouteInputWithoutSourceReturnsFalse(t *testing.T) {
	h := NewPtyHub()

	if ok := h.RouteInput("sess-a", []byte("input")); ok {
		t.Fatal("expected RouteInput to report no delivery when no source is connected")
	}
}

func TestPtyHubRouteInputIsolatesSessions(t *testing.T) {
	h := NewPtyHub()

	deliver, unregister := h.SetSource("sess-a")
	defer unregister()

	if ok := h.RouteInput("sess-b", []byte("input")); ok {
		t.Fatal("expected RouteInput for other session to report no delivery")
	}

	select {
	case msg := <-deliver:
		t.Fatalf("got unexpected message routed from another session: %q", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestPtyHubSetSourceReplacesExisting(t *testing.T) {
	h := NewPtyHub()

	first, unregisterFirst := h.SetSource("sess-a")
	_ = unregisterFirst
	second, unregisterSecond := h.SetSource("sess-a")
	defer unregisterSecond()

	if ok := h.RouteInput("sess-a", []byte("input")); !ok {
		t.Fatal("expected RouteInput to report delivery to the replacement source")
	}

	select {
	case msg := <-first:
		t.Fatalf("stale source should not receive routed input, got %q", msg)
	case <-time.After(50 * time.Millisecond):
	}

	select {
	case msg := <-second:
		if string(msg) != "input" {
			t.Fatalf("got %q, want input", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for routed input on replacement source")
	}
}

func TestPtyHubUnregisterStaleSourceDoesNotEvictReplacement(t *testing.T) {
	h := NewPtyHub()

	_, unregisterFirst := h.SetSource("sess-a")
	_, unregisterSecond := h.SetSource("sess-a")
	defer unregisterSecond()

	unregisterFirst()

	if ok := h.RouteInput("sess-a", []byte("input")); !ok {
		t.Fatal("expected the replacement source to still be registered after the stale one unregisters")
	}
}

func TestPtyHubConcurrentAccess(t *testing.T) {
	h := NewPtyHub()

	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			ch, unsub := h.Subscribe("sess-a")
			defer unsub()
			select {
			case <-ch:
			case <-time.After(time.Second):
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < 20; i++ {
		go func() {
			_, unregister := h.SetSource("sess-a")
			defer unregister()
			h.RouteInput("sess-a", []byte("input"))
		}()
	}

	for i := 0; i < 20; i++ {
		go h.Broadcast("sess-a", []byte("output"))
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}
