package stream

import "sync"

const presenceBuffer = 32

type PresenceHub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan []byte]struct{}
}

func NewPresenceHub() *PresenceHub {
	return &PresenceHub{subscribers: map[string]map[chan []byte]struct{}{}}
}

func (h *PresenceHub) Subscribe(sessionID string) (chan []byte, func()) {
	ch := make(chan []byte, presenceBuffer)

	h.mu.Lock()
	if h.subscribers[sessionID] == nil {
		h.subscribers[sessionID] = map[chan []byte]struct{}{}
	}
	h.subscribers[sessionID][ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		if _, ok := h.subscribers[sessionID][ch]; ok {
			delete(h.subscribers[sessionID], ch)
			close(ch)
		}
	}

	return ch, unsubscribe
}

func (h *PresenceHub) Broadcast(sessionID string, except chan []byte, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subscribers[sessionID] {
		if ch == except {
			continue
		}
		select {
		case ch <- msg:
		default:
		}
	}
}
