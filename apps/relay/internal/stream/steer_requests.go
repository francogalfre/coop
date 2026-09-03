package stream

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
)

type PendingSteerRequest struct {
	RequestID string
	Actor     auth.Actor
	Text      string
}

const steerRequestCap = 16

type SteerRequestRegistry struct {
	pool    *db.Pool
	mu      sync.Mutex
	pending map[string]map[string]PendingSteerRequest
	order   map[string][]string
}

func NewSteerRequestRegistry(pool *db.Pool) *SteerRequestRegistry {
	return &SteerRequestRegistry{
		pool:    pool,
		pending: map[string]map[string]PendingSteerRequest{},
		order:   map[string][]string{},
	}
}

func (r *SteerRequestRegistry) Put(ctx context.Context, sessionID string, req PendingSteerRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reqs := r.pending[sessionID]
	if reqs == nil {
		reqs = map[string]PendingSteerRequest{}
		r.pending[sessionID] = reqs
	}

	order := r.order[sessionID]
	if len(order) >= steerRequestCap {
		oldest := order[0]
		order = order[1:]
		delete(reqs, oldest)
		log.Printf("coop: steer request queue full for session %s, dropped oldest request %s", sessionID, oldest)
	}

	reqs[req.RequestID] = req
	r.order[sessionID] = append(order, req.RequestID)

	if r.pool == nil {
		return nil
	}

	if err := r.pool.CreateSteerRequest(ctx, req.RequestID, sessionID, req.Actor.UserID, req.Actor.DisplayName, req.Actor.AvatarURL, req.Text); err != nil {
		return fmt.Errorf("steer requests: put: %w", err)
	}

	if _, err := r.pool.EvictOldestSteerRequests(ctx, sessionID, steerRequestCap); err != nil {
		return fmt.Errorf("steer requests: put: evict: %w", err)
	}

	return nil
}

func (r *SteerRequestRegistry) Take(ctx context.Context, sessionID, requestID string) (PendingSteerRequest, bool, error) {
	req, ok := r.takeCached(sessionID, requestID)
	if ok {
		if r.pool != nil {
			if err := r.pool.DeleteSteerRequest(ctx, requestID); err != nil {
				return PendingSteerRequest{}, false, fmt.Errorf("steer requests: take: %w", err)
			}
		}

		return req, true, nil
	}

	if r.pool == nil {
		return PendingSteerRequest{}, false, nil
	}

	// A cache miss (e.g. after a restart) falls back to Postgres so the request stays resolvable.
	row, err := r.pool.GetSteerRequest(ctx, sessionID, requestID)
	if err != nil {
		if db.IsNotFound(err) {
			return PendingSteerRequest{}, false, nil
		}

		return PendingSteerRequest{}, false, fmt.Errorf("steer requests: take: %w", err)
	}

	if err := r.pool.DeleteSteerRequest(ctx, requestID); err != nil {
		return PendingSteerRequest{}, false, fmt.Errorf("steer requests: take: %w", err)
	}

	return PendingSteerRequest{
		RequestID: row.ID,
		Actor:     auth.Actor{UserID: row.ActorID, DisplayName: row.ActorDisplayName, AvatarURL: row.ActorAvatarURL},
		Text:      row.Text,
	}, true, nil
}

func (r *SteerRequestRegistry) takeCached(sessionID, requestID string) (PendingSteerRequest, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reqs := r.pending[sessionID]
	req, ok := reqs[requestID]
	if !ok {
		return PendingSteerRequest{}, false
	}

	delete(reqs, requestID)

	order := r.order[sessionID]
	for i, id := range order {
		if id == requestID {
			r.order[sessionID] = append(order[:i], order[i+1:]...)
			break
		}
	}

	return req, true
}
