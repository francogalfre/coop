package presence

import (
	"fmt"
	"sync"
	"time"
)

type Registry struct {
	mu       sync.RWMutex
	sessions map[string]*sessionState
	touches  map[string]map[string]map[string]touch
}

type sessionState struct {
	Repo      string
	Owner     string
	StartedAt time.Time
	Ended     bool
	EndedAt   time.Time
}

type touch struct {
	Mode string
	At   time.Time
}

func New() *Registry {
	return &Registry{
		sessions: map[string]*sessionState{},
		touches:  map[string]map[string]map[string]touch{},
	}
}

func (r *Registry) SessionStarted(sessionID, repo, owner string, at time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[sessionID] = &sessionState{
		Repo:      repo,
		Owner:     owner,
		StartedAt: at,
	}
}

func (r *Registry) SessionEnded(sessionID string, at time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return
	}

	session.Ended = true
	session.EndedAt = at
}

func (r *Registry) FileTouched(sessionID, path, mode string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return fmt.Errorf("file touched: unknown session %q", sessionID)
	}

	byPath, ok := r.touches[session.Repo]
	if !ok {
		byPath = map[string]map[string]touch{}
		r.touches[session.Repo] = byPath
	}

	bySession, ok := byPath[path]
	if !ok {
		bySession = map[string]touch{}
		byPath[path] = bySession
	}

	bySession[sessionID] = touch{Mode: mode, At: at}

	return nil
}

func (r *Registry) Sweep(now time.Time, maxAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, session := range r.sessions {
		if session.Ended && now.Sub(session.EndedAt) > maxAge {
			delete(r.sessions, id)
		}
	}

	for repo, byPath := range r.touches {
		for path, bySession := range byPath {
			for sessionID, t := range bySession {
				if now.Sub(t.At) > maxAge {
					delete(bySession, sessionID)
				}
			}

			if len(bySession) == 0 {
				delete(byPath, path)
			}
		}

		if len(byPath) == 0 {
			delete(r.touches, repo)
		}
	}
}
