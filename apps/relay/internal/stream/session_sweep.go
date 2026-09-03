package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
)

func SweepStaleSessions(ctx context.Context, pool *db.Pool, store *Store, idleSince time.Time) ([]string, error) {
	candidates, err := pool.StaleSessionIDs(ctx, idleSince)
	if err != nil {
		return nil, fmt.Errorf("stream: sweep stale sessions: %w", err)
	}

	var ended []string

	for _, id := range candidates {
		didEnd, err := pool.EndSessionIfStale(ctx, id, idleSince)
		if err != nil {
			// Leave the remaining candidates untouched so the next tick retries them.
			return ended, fmt.Errorf("stream: sweep stale sessions: %w", err)
		}
		if !didEnd {
			continue
		}

		ended = append(ended, id)

		// The row status is authoritative; the synthetic event only nudges live
		// viewers, so a failure here is logged rather than allowed to abort the sweep.
		if err := emitSessionEnd(ctx, pool, store, id); err != nil {
			log.Printf("stream: sweep stale sessions: emit session.end for %s: %v", id, err)
		}
	}

	return ended, nil
}

func emitSessionEnd(ctx context.Context, pool *db.Pool, store *Store, sessionID string) error {
	body, err := json.Marshal(map[string]any{
		"v":          1,
		"session_id": sessionID,
		"seq":        0,
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"type":       "session.end",
	})
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	dbEvent, err := pool.AppendEvent(ctx, sessionID, body)
	if err != nil {
		return fmt.Errorf("append: %w", err)
	}

	if _, err := store.AppendWithSeq(sessionID, dbEvent.Seq, json.RawMessage(body)); err != nil {
		return fmt.Errorf("broadcast: %w", err)
	}

	return nil
}
