package wsapi

import (
	"context"
	"net/http"

	"github.com/coder/websocket"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const backfillLimit = 200
const maxSessionFrameBytes = 64 * 1024

func NewSessionStreamHandler(pool *db.Pool, store *stream.Store, hub *stream.PresenceHub, webOrigins []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}
		name := actor.DisplayName

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: webOrigins})
		if err != nil {
			return
		}
		defer conn.CloseNow()
		conn.SetReadLimit(maxSessionFrameBytes)

		ctx := r.Context()

		live, unsubscribeEvents := store.Subscribe(sessionID)
		defer unsubscribeEvents()

		presenceCh, unsubscribePresence := hub.Subscribe(sessionID)
		defer unsubscribePresence()

		hub.AddViewer(sessionID, name)
		defer hub.RemoveViewer(sessionID, name)

		broadcastPresence(hub, sessionID, presenceCh, name, "human.join")
		defer broadcastPresence(hub, sessionID, presenceCh, name, "human.leave")

		incoming := make(chan clientPresenceMessage)
		go readClientFrames(ctx, conn, incoming)

		lastSeq := 0
		for _, event := range backfillEvents(ctx, pool, store, sessionID) {
			if err := conn.Write(ctx, websocket.MessageText, event.Data); err != nil {
				return
			}
			lastSeq = event.Seq
		}

		for {
			select {
			case <-ctx.Done():
				conn.Close(websocket.StatusNormalClosure, "")
				return
			case event, ok := <-live:
				if !ok {
					return
				}
				if event.Seq <= lastSeq {
					continue
				}
				if err := conn.Write(ctx, websocket.MessageText, event.Data); err != nil {
					return
				}
				lastSeq = event.Seq
			case msg, ok := <-incoming:
				if !ok {
					return
				}
				broadcastTyping(hub, sessionID, presenceCh, name, msg.Active)
			case frame, ok := <-presenceCh:
				if !ok {
					return
				}
				if err := conn.Write(ctx, websocket.MessageText, frame); err != nil {
					return
				}
			}
		}
	}
}

// Postgres is tried first since the in-memory ring buffer is gone on restart; an empty/failed read falls back to it.
func backfillEvents(ctx context.Context, pool *db.Pool, store *stream.Store, sessionID string) []stream.Event {
	if pool != nil {
		if rows, err := pool.RecentEvents(ctx, sessionID, backfillLimit); err == nil && len(rows) > 0 {
			events := make([]stream.Event, 0, len(rows))
			for _, row := range rows {
				data, err := stream.StampSeq(row.Data, row.Seq)
				if err != nil {
					continue
				}
				events = append(events, stream.Event{Seq: row.Seq, Data: data})
			}
			return events
		}
	}

	return store.Since(sessionID, 0)
}
