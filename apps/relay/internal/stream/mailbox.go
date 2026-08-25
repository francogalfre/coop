package stream

import "sync"

type SteerMessage struct {
	From string
	Text string
}

type Mailbox struct {
	mu      sync.Mutex
	pending map[string]SteerMessage
}

func NewMailbox() *Mailbox {
	return &Mailbox{pending: map[string]SteerMessage{}}
}

func (m *Mailbox) Put(sessionID string, msg SteerMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pending[sessionID] = msg
}

func (m *Mailbox) Take(sessionID string) (SteerMessage, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg, ok := m.pending[sessionID]
	if ok {
		delete(m.pending, sessionID)
	}

	return msg, ok
}
