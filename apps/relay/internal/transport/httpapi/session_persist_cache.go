package httpapi

import (
	"context"
	"sync"

	"github.com/francogalfre/coop/apps/relay/internal/db"
)

const sessionPersistCacheCap = 4096

type sessionPersistCache struct {
	mu sync.RWMutex
	m  map[string]bool
}

func newSessionPersistCache() *sessionPersistCache {
	return &sessionPersistCache{m: map[string]bool{}}
}

func (c *sessionPersistCache) markPersisted(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.m) >= sessionPersistCacheCap {
		c.m = map[string]bool{}
	}
	c.m[sessionID] = true
}

func (c *sessionPersistCache) isPersisted(sessionID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m[sessionID]
}

// A miss falls back to Postgres so a restart (which clears the cache) never silently stops persisting.
func (c *sessionPersistCache) check(ctx context.Context, pool *db.Pool, sessionID string) bool {
	if c.isPersisted(sessionID) {
		return true
	}

	if pool == nil {
		return false
	}

	if _, err := pool.GetAgentSession(ctx, sessionID); err != nil {
		return false
	}

	c.markPersisted(sessionID)
	return true
}
