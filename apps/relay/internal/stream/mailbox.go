package stream

import (
	"log"
	"sync"
)

type SteerMessage struct {
	From string
	Text string
}

const mailboxCap = 32

type Mailbox struct {
	mu      sync.Mutex
	pending map[string][]SteerMessage
}

func NewMailbox() *Mailbox {
	return &Mailbox{pending: map[string][]SteerMessage{}}
}

func (m *Mailbox) Put(sessionID string, msg SteerMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue := m.pending[sessionID]
	if len(queue) >= mailboxCap {
		log.Printf("coop: mailbox full for session %s, dropped oldest message", sessionID)
		queue = queue[1:]
	}

	m.pending[sessionID] = append(queue, msg)
}

func (m *Mailbox) Take(sessionID string) (SteerMessage, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue := m.pending[sessionID]
	if len(queue) == 0 {
		return SteerMessage{}, false
	}

	msg := queue[0]
	m.pending[sessionID] = queue[1:]

	return msg, true
}

func (m *Mailbox) Depth(sessionID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.pending[sessionID])
}
