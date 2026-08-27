package stream

import (
	"log"
	"sync"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
)

type PendingSteerRequest struct {
	RequestID string
	Actor     auth.Actor
	Text      string
}

const steerRequestCap = 16

type SteerRequestRegistry struct {
	mu      sync.Mutex
	pending map[string]map[string]PendingSteerRequest
	order   map[string][]string
}

func NewSteerRequestRegistry() *SteerRequestRegistry {
	return &SteerRequestRegistry{
		pending: map[string]map[string]PendingSteerRequest{},
		order:   map[string][]string{},
	}
}

func (r *SteerRequestRegistry) Put(sessionID string, req PendingSteerRequest) {
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
}

func (r *SteerRequestRegistry) Take(sessionID, requestID string) (PendingSteerRequest, bool) {
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
