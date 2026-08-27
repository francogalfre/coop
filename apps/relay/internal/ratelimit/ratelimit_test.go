package ratelimit

import (
	"fmt"
	"testing"
)

func TestAllowBoundsBucketMapUnderDiverseIPs(t *testing.T) {
	l := New(1.0, 5.0)

	for i := 0; i < maxBuckets*3; i++ {
		l.Allow(fmt.Sprintf("ip-%d", i))
	}

	l.mu.Lock()
	got := len(l.buckets)
	l.mu.Unlock()

	if got > maxBuckets {
		t.Fatalf("got %d buckets, want at most %d", got, maxBuckets)
	}
}

func TestAllowStillRateLimitsAfterEviction(t *testing.T) {
	l := New(1.0, 1.0)

	for i := 0; i < maxBuckets+10; i++ {
		l.Allow(fmt.Sprintf("filler-%d", i))
	}

	if !l.Allow("client") {
		t.Fatal("expected first request from a fresh key to be allowed")
	}
	if l.Allow("client") {
		t.Fatal("expected second immediate request to be denied by the rate limit")
	}
}
