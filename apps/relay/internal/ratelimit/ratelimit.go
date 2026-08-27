package ratelimit

import (
	"sync"
	"time"
)

const maxBuckets = 8192

type bucket struct {
	tokens   float64
	lastFill time.Time
}

type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64
	capacity float64
}

func New(ratePerSecond float64, capacity float64) *Limiter {
	return &Limiter{
		buckets:  map[string]*bucket{},
		rate:     ratePerSecond,
		capacity: capacity,
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	b, ok := l.buckets[key]
	if !ok {
		if len(l.buckets) >= maxBuckets {
			l.buckets = map[string]*bucket{}
		}

		b = &bucket{tokens: l.capacity, lastFill: now}
		l.buckets[key] = b
	}

	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens = min(l.capacity, b.tokens+elapsed*l.rate)
	b.lastFill = now

	if b.tokens < 1 {
		return false
	}

	b.tokens--

	return true
}
