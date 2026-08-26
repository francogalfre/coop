package stream

import "testing"

func TestMailboxPutThenTake(t *testing.T) {
	m := NewMailbox()

	m.Put("sess-a", SteerMessage{From: "Alice", Text: "try the other branch"})

	msg, ok := m.Take("sess-a")
	if !ok {
		t.Fatal("expected a pending message")
	}
	if msg.From != "Alice" || msg.Text != "try the other branch" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestMailboxTakeIsOneShot(t *testing.T) {
	m := NewMailbox()

	m.Put("sess-a", SteerMessage{From: "Alice", Text: "hello"})
	m.Take("sess-a")

	_, ok := m.Take("sess-a")
	if ok {
		t.Fatal("expected second take to find nothing")
	}
}

func TestMailboxTakeUnknownSession(t *testing.T) {
	m := NewMailbox()

	_, ok := m.Take("sess-ghost")
	if ok {
		t.Fatal("expected no message for unknown session")
	}
}

func TestMailboxPutQueuesInOrderInsteadOfOverwriting(t *testing.T) {
	m := NewMailbox()

	m.Put("sess-a", SteerMessage{From: "Alice", Text: "first"})
	m.Put("sess-a", SteerMessage{From: "Bob", Text: "second"})

	first, ok := m.Take("sess-a")
	if !ok {
		t.Fatal("expected first pending message")
	}
	if first.From != "Alice" || first.Text != "first" {
		t.Fatalf("unexpected first message: %+v", first)
	}

	second, ok := m.Take("sess-a")
	if !ok {
		t.Fatal("expected second pending message")
	}
	if second.From != "Bob" || second.Text != "second" {
		t.Fatalf("unexpected second message: %+v", second)
	}
}

func TestMailboxDepth(t *testing.T) {
	m := NewMailbox()

	if got := m.Depth("sess-a"); got != 0 {
		t.Fatalf("got depth %d, want 0", got)
	}

	m.Put("sess-a", SteerMessage{From: "Alice", Text: "first"})
	m.Put("sess-a", SteerMessage{From: "Bob", Text: "second"})

	if got := m.Depth("sess-a"); got != 2 {
		t.Fatalf("got depth %d, want 2", got)
	}

	m.Take("sess-a")

	if got := m.Depth("sess-a"); got != 1 {
		t.Fatalf("got depth %d, want 1", got)
	}
}

func TestMailboxPutDropsOldestBeyondCap(t *testing.T) {
	m := NewMailbox()

	for i := 0; i < mailboxCap+1; i++ {
		m.Put("sess-a", SteerMessage{From: "Alice", Text: string(rune('a' + i))})
	}

	if got := m.Depth("sess-a"); got != mailboxCap {
		t.Fatalf("got depth %d, want %d", got, mailboxCap)
	}

	first, ok := m.Take("sess-a")
	if !ok {
		t.Fatal("expected a pending message")
	}
	if first.Text != string(rune('a'+1)) {
		t.Fatalf("expected oldest message dropped, got %+v", first)
	}
}
