package stream

import (
	"testing"
	"time"
)

func TestPresenceHubBroadcastReceivedByOthers(t *testing.T) {
	h := NewPresenceHub()

	a, unsubA := h.Subscribe("sess-a")
	defer unsubA()
	b, unsubB := h.Subscribe("sess-a")
	defer unsubB()

	h.Broadcast("sess-a", a, []byte("hello"))

	select {
	case msg := <-b:
		if string(msg) != "hello" {
			t.Fatalf("got %q, want hello", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broadcast")
	}
}

func TestPresenceHubBroadcastExcludesSender(t *testing.T) {
	h := NewPresenceHub()

	a, unsubA := h.Subscribe("sess-a")
	defer unsubA()
	_, unsubB := h.Subscribe("sess-a")
	defer unsubB()

	h.Broadcast("sess-a", a, []byte("hello"))

	select {
	case msg := <-a:
		t.Fatalf("sender should not receive its own broadcast, got %q", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestPresenceHubBroadcastDoesNotReceiveOtherSessions(t *testing.T) {
	h := NewPresenceHub()

	ch, unsub := h.Subscribe("sess-a")
	defer unsub()

	h.Broadcast("sess-b", nil, []byte("hello"))

	select {
	case msg := <-ch:
		t.Fatalf("got unexpected message for other session: %q", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestPresenceHubUnsubscribeClosesChannel(t *testing.T) {
	h := NewPresenceHub()

	ch, unsub := h.Subscribe("sess-a")
	unsub()

	_, open := <-ch
	if open {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}

func TestPresenceHubConcurrentSubscribeAndBroadcast(t *testing.T) {
	h := NewPresenceHub()

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
		go h.Broadcast("sess-a", nil, []byte("hello"))
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestPresenceHubViewersTracksAddAndRemove(t *testing.T) {
	h := NewPresenceHub()

	if got := h.Viewers("sess-a"); len(got) != 0 {
		t.Fatalf("got %v, want no viewers before any join", got)
	}

	h.AddViewer("sess-a", "Alice")
	h.AddViewer("sess-a", "Bob")

	got := h.Viewers("sess-a")
	if len(got) != 2 {
		t.Fatalf("got %v, want 2 viewers", got)
	}

	h.RemoveViewer("sess-a", "Alice")

	got = h.Viewers("sess-a")
	if len(got) != 1 || got[0] != "Bob" {
		t.Fatalf("got %v, want only Bob left", got)
	}

	h.RemoveViewer("sess-a", "Bob")

	if got := h.Viewers("sess-a"); len(got) != 0 {
		t.Fatalf("got %v, want no viewers after everyone left", got)
	}
}

func TestPresenceHubViewersHandlesSameNameTwice(t *testing.T) {
	h := NewPresenceHub()

	h.AddViewer("sess-a", "Alice")
	h.AddViewer("sess-a", "Alice")

	if got := h.Viewers("sess-a"); len(got) != 1 {
		t.Fatalf("got %v, want one entry for two tabs with the same name", got)
	}

	h.RemoveViewer("sess-a", "Alice")

	if got := h.Viewers("sess-a"); len(got) != 1 {
		t.Fatalf("got %v, want Alice still present after closing only one tab", got)
	}

	h.RemoveViewer("sess-a", "Alice")

	if got := h.Viewers("sess-a"); len(got) != 0 {
		t.Fatalf("got %v, want no viewers after closing the second tab", got)
	}
}

func TestPresenceHubViewersIgnoresEmptyName(t *testing.T) {
	h := NewPresenceHub()

	h.AddViewer("sess-a", "")

	if got := h.Viewers("sess-a"); len(got) != 0 {
		t.Fatalf("got %v, want empty names ignored", got)
	}
}
