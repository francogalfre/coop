package stream

import "sync"

type TakeoverState struct {
	Active bool
	ByID   string
	By     string
}

type TakeoverRegistry struct {
	mu    sync.RWMutex
	state map[string]TakeoverState
}

func NewTakeoverRegistry() *TakeoverRegistry {
	return &TakeoverRegistry{state: map[string]TakeoverState{}}
}

func (r *TakeoverRegistry) Get(sessionID string) TakeoverState {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.state[sessionID]
}

// Storing a released state as a deleted map entry keeps Get's zero-value default meaningful.
func (r *TakeoverRegistry) Set(sessionID string, state TakeoverState) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !state.Active {
		delete(r.state, sessionID)
		return
	}

	r.state[sessionID] = state
}
