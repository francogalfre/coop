package stream

import "sync"

const ptyBuffer = 256

type ptySource struct {
	deliver chan []byte
}

type PtyHub struct {
	mu          sync.RWMutex
	sources     map[string]*ptySource
	subscribers map[string]map[chan []byte]struct{}
}

func NewPtyHub() *PtyHub {
	return &PtyHub{
		sources:     map[string]*ptySource{},
		subscribers: map[string]map[chan []byte]struct{}{},
	}
}

// A reconnecting CLI is the common case, so a new source silently replaces a stale one rather than erroring.
func (h *PtyHub) SetSource(sessionID string) (chan []byte, func()) {
	src := &ptySource{deliver: make(chan []byte, ptyBuffer)}

	h.mu.Lock()
	h.sources[sessionID] = src
	h.mu.Unlock()

	unregister := func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		if h.sources[sessionID] == src {
			delete(h.sources, sessionID)
		}
	}

	return src.deliver, unregister
}

func (h *PtyHub) Subscribe(sessionID string) (chan []byte, func()) {
	ch := make(chan []byte, ptyBuffer)

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

func (h *PtyHub) Broadcast(sessionID string, frame []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subscribers[sessionID] {
		select {
		case ch <- frame:
		default:
		}
	}
}

func (h *PtyHub) RouteInput(sessionID string, frame []byte) bool {
	h.mu.RLock()
	src, ok := h.sources[sessionID]
	h.mu.RUnlock()

	if !ok {
		return false
	}

	select {
	case src.deliver <- frame:
		return true
	default:
		return false
	}
}
