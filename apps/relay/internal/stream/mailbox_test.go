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

func TestMailboxPutReplacesPending(t *testing.T) {
	m := NewMailbox()

	m.Put("sess-a", SteerMessage{From: "Alice", Text: "first"})
	m.Put("sess-a", SteerMessage{From: "Bob", Text: "second"})

	msg, ok := m.Take("sess-a")
	if !ok {
		t.Fatal("expected a pending message")
	}
	if msg.From != "Bob" || msg.Text != "second" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}
