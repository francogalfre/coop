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

// Shared with history reads so a replayed event matches the wire shape the live WS broadcast used.
func StampSeq(event json.RawMessage, seq int) (json.RawMessage, error) {
	var fields map[string]any
	if err := json.Unmarshal(event, &fields); err != nil {
		return nil, fmt.Errorf("stamp seq: unmarshal event: %w", err)
	}

	fields["seq"] = seq

	data, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("stamp seq: marshal event: %w", err)
	}

	return data, nil
}

func (s *Store) Append(sessionID string, event json.RawMessage) (Event, error) {
	s.mu.Lock()

	buf, ok := s.sessions[sessionID]
	if !ok {
		buf = &sessionBuffer{}
		s.sessions[sessionID] = buf
	}
	buf.nextSeq++
	seq := buf.nextSeq

	data, err := StampSeq(event, seq)
	if err != nil {
		s.mu.Unlock()
		return Event{}, fmt.Errorf("append: %w", err)
	}

	recorded := Event{Seq: seq, Data: data}

	buf.events = append(buf.events, recorded)
	if len(buf.events) > bufferSize {
		buf.events = buf.events[len(buf.events)-bufferSize:]
	}
	s.mu.Unlock()

	s.publish(sessionID, recorded)

	return recorded, nil
}

func (s *Store) AppendWithSeq(sessionID string, seq int, event json.RawMessage) (Event, error) {
	data, err := StampSeq(event, seq)
	if err != nil {
		return Event{}, fmt.Errorf("append with seq: %w", err)
	}

	recorded := Event{Seq: seq, Data: data}

	s.mu.Lock()
	buf, ok := s.sessions[sessionID]
	if !ok {
		buf = &sessionBuffer{}
		s.sessions[sessionID] = buf
	}
	if seq > buf.nextSeq {
		buf.nextSeq = seq
	}

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
