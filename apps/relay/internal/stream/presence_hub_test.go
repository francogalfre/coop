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
