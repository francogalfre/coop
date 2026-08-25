package stream

const subscriberBuffer = 32

func (s *Store) Subscribe(sessionID string) (<-chan Event, func()) {
	ch := make(chan Event, subscriberBuffer)

	s.mu.Lock()
	if s.subscribers[sessionID] == nil {
		s.subscribers[sessionID] = map[chan Event]struct{}{}
	}
	s.subscribers[sessionID][ch] = struct{}{}
	s.mu.Unlock()

	unsubscribe := func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if _, ok := s.subscribers[sessionID][ch]; ok {
			delete(s.subscribers[sessionID], ch)
			close(ch)
		}
	}

	return ch, unsubscribe
}

func (s *Store) publish(sessionID string, event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for ch := range s.subscribers[sessionID] {
		select {
		case ch <- event:
		default:
		}
	}
}
