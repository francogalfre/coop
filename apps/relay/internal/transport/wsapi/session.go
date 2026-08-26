package wsapi

import (
	"net/http"

	"github.com/coder/websocket"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func NewSessionStreamHandler(store *stream.Store, hub *stream.PresenceHub, webOrigins []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")
		name := r.URL.Query().Get("name")

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: webOrigins})
		if err != nil {
			return
		}
		defer conn.CloseNow()

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
