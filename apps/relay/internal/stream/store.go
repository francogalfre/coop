package stream

import (
	"encoding/json"
	"fmt"
	"sync"
)

const bufferSize = 500

type Store struct {
	mu          sync.RWMutex
	sessions    map[string]*sessionBuffer
	subscribers map[string]map[chan Event]struct{}
}

type sessionBuffer struct {
	events  []Event
	nextSeq int
}

func New() *Store {
	return &Store{
		sessions:    map[string]*sessionBuffer{},
		subscribers: map[string]map[chan Event]struct{}{},
	}
}

func (s *Store) Append(sessionID string, event json.RawMessage) (Event, error) {
	var fields map[string]any
	if err := json.Unmarshal(event, &fields); err != nil {
		return Event{}, fmt.Errorf("append: unmarshal event: %w", err)
	}

	s.mu.Lock()
	buf, ok := s.sessions[sessionID]
	if !ok {
		buf = &sessionBuffer{}
		s.sessions[sessionID] = buf
	}

	buf.nextSeq++
	fields["seq"] = buf.nextSeq

	data, err := json.Marshal(fields)
	if err != nil {
		s.mu.Unlock()
		return Event{}, fmt.Errorf("append: marshal event: %w", err)
	}

	recorded := Event{Seq: buf.nextSeq, Data: data}

	buf.events = append(buf.events, recorded)
	if len(buf.events) > bufferSize {
		buf.events = buf.events[len(buf.events)-bufferSize:]
	}
	s.mu.Unlock()

	s.publish(sessionID, recorded)

	return recorded, nil
}

func (s *Store) Since(sessionID string, afterSeq int) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buf, ok := s.sessions[sessionID]
	if !ok {
		return []Event{}
	}

	result := make([]Event, 0, len(buf.events))
	for _, e := range buf.events {
		if e.Seq > afterSeq {
			result = append(result, e)
		}
	}

	return result
}
