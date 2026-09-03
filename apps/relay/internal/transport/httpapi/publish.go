package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

// publishEvent persists the event and mirrors it into the in-memory store in one call.
func publishEvent(ctx context.Context, pool *db.Pool, store *stream.Store, sessionID string, envelope json.RawMessage) (int, error) {
	dbEvent, err := pool.AppendEvent(ctx, sessionID, envelope)
	if err != nil {
		return 0, fmt.Errorf("publish event: %w", err)
	}

	if _, err := store.AppendWithSeq(sessionID, dbEvent.Seq, envelope); err != nil {
		log.Printf("coop: failed to append event to in-memory store for session %s: %v", sessionID, err)
	}

	return dbEvent.Seq, nil
}
