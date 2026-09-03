package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
)

// The synthetic session.end is what makes an open browser drop the session from
// "Live"; ending the row alone only fixes a fresh page load.
func SweepStaleSessions(ctx context.Context, pool *db.Pool, store *Store, idleSince time.Time) ([]string, error) {
	ended, err := pool.EndStaleSessions(ctx, idleSince)
	if err != nil {
		return nil, fmt.Errorf("stream: sweep stale sessions: %w", err)
	}

	ts := time.Now().UTC().Format(time.RFC3339)

	for _, id := range ended {
		body, err := json.Marshal(map[string]any{
			"v":          1,
			"session_id": id,
			"seq":        0,
			"ts":         ts,
			"type":       "session.end",
		})
		if err != nil {
			return ended, fmt.Errorf("stream: sweep stale sessions: marshal %s: %w", id, err)
		}

		dbEvent, err := pool.AppendEvent(ctx, id, body)
		if err != nil {
			return ended, fmt.Errorf("stream: sweep stale sessions: append %s: %w", id, err)
		}

		if _, err := store.AppendWithSeq(id, dbEvent.Seq, json.RawMessage(body)); err != nil {
			return ended, fmt.Errorf("stream: sweep stale sessions: broadcast %s: %w", id, err)
		}
	}

	return ended, nil
}
