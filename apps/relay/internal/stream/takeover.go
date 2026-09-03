package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/francogalfre/coop/apps/relay/internal/db"
)

type TakeoverState struct {
	Active bool
	ByID   string
	By     string
}

type TakeoverRegistry struct {
	pool  *db.Pool
	mu    sync.RWMutex
	state map[string]TakeoverState
}

func NewTakeoverRegistry(pool *db.Pool) *TakeoverRegistry {
	return &TakeoverRegistry{pool: pool, state: map[string]TakeoverState{}}
}

// A cache miss falls back to Postgres so a restart (which clears the cache) never loses a live takeover.
func (r *TakeoverRegistry) Get(ctx context.Context, sessionID string) (TakeoverState, error) {
	r.mu.RLock()
	state, cached := r.state[sessionID]
	r.mu.RUnlock()

	if cached {
		return state, nil
	}

	if r.pool == nil {
		return TakeoverState{}, nil
	}

	row, err := r.pool.GetTakeover(ctx, sessionID)
	if err != nil {
		if db.IsNotFound(err) {
			return TakeoverState{}, nil
		}

		return TakeoverState{}, fmt.Errorf("takeover: get: %w", err)
	}

	state = TakeoverState{Active: true, ByID: row.ActorID, By: row.ActorDisplayName}

	r.mu.Lock()
	r.state[sessionID] = state
	r.mu.Unlock()

	return state, nil
}

// Storing a released state as a deleted map entry keeps Get's zero-value default meaningful.
func (r *TakeoverRegistry) Set(ctx context.Context, sessionID string, state TakeoverState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !state.Active {
		delete(r.state, sessionID)

		if r.pool == nil {
			return nil
		}

		if err := r.pool.ClearTakeover(ctx, sessionID); err != nil {
			return fmt.Errorf("takeover: clear: %w", err)
		}

		return nil
	}

	if r.pool != nil {
		if err := r.pool.SetTakeover(ctx, sessionID, state.ByID, state.By); err != nil {
			return fmt.Errorf("takeover: set: %w", err)
		}
	}

	r.state[sessionID] = state

	return nil
}
