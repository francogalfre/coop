package stream

import (
	"context"
	"sync"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
)

type QuestionAnswer struct {
	Text  string
	Actor auth.Actor
}

type pendingQuestion struct {
	answer  *QuestionAnswer
	waiters []chan struct{}
}

// QuestionRegistry backs agent.asked_team / human.answered: in-memory only, since a relay restart mid-question is indistinguishable from nobody answering.
type QuestionRegistry struct {
	mu        sync.Mutex
	questions map[string]*pendingQuestion
}

func NewQuestionRegistry() *QuestionRegistry {
	return &QuestionRegistry{questions: map[string]*pendingQuestion{}}
}

func (r *QuestionRegistry) Open(questionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.questions[questionID] = &pendingQuestion{}
}

func (r *QuestionRegistry) Answer(questionID string, answer QuestionAnswer) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	q, ok := r.questions[questionID]
	if !ok || q.answer != nil {
		return false
	}

	q.answer = &answer
	for _, waiter := range q.waiters {
		close(waiter)
	}
	q.waiters = nil

	return true
}

// Wait blocks until the question is answered, ctx is cancelled, or wait elapses.
func (r *QuestionRegistry) Wait(ctx context.Context, questionID string, wait time.Duration) (QuestionAnswer, bool, bool) {
	r.mu.Lock()
	q, ok := r.questions[questionID]
	if !ok {
		r.mu.Unlock()
		return QuestionAnswer{}, false, false
	}

	if q.answer != nil {
		answer := *q.answer
		r.mu.Unlock()
		return answer, true, true
	}

	ch := make(chan struct{})
	q.waiters = append(q.waiters, ch)
	r.mu.Unlock()

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ch:
		r.mu.Lock()
		defer r.mu.Unlock()
		if q.answer == nil {
			return QuestionAnswer{}, false, true
		}
		return *q.answer, true, true
	case <-timer.C:
		return QuestionAnswer{}, false, true
	case <-ctx.Done():
		return QuestionAnswer{}, false, true
	}
}
