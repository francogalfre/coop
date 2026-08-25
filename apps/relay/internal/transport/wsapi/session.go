package wsapi

import (
	"net/http"

	"github.com/coder/websocket"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func NewSessionStreamHandler(store *stream.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()

		ctx := r.Context()

		live, unsubscribe := store.Subscribe(sessionID)
		defer unsubscribe()

		lastSeq := 0
		for _, event := range store.Since(sessionID, 0) {
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
			}
		}
	}
}
