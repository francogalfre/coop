package wsapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coder/websocket"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const maxPtyFrameBytes = 256 * 1024

type ptyFrameType struct {
	Type string `json:"type"`
}

func NewPtySessionHandler(hub *stream.PtyHub, takeover *stream.TakeoverRegistry, pool *db.Pool, webOrigins []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		isSource := isPtySource(r, pool, sessionID, actor.UserID)

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: webOrigins})
		if err != nil {
			return
		}
		defer conn.CloseNow()
		conn.SetReadLimit(maxPtyFrameBytes)

		ctx := r.Context()

		if isSource {
			runPtySource(ctx, conn, hub, sessionID)
			return
		}

		runPtyViewer(ctx, conn, hub, takeover, sessionID, actor.UserID)
	}
}

// Mirrors the CLI-vs-browser dispatch RequireAnyIdentity already does; owning the session is what actually makes it the source.
func isPtySource(r *http.Request, pool *db.Pool, sessionID, userID string) bool {
	if r.Header.Get("Authorization") == "" {
		return false
	}

	sess, err := pool.GetAgentSession(r.Context(), sessionID)
	if err != nil {
		return false
	}

	return sess.OwnerID == userID
}

func runPtySource(ctx context.Context, conn *websocket.Conn, hub *stream.PtyHub, sessionID string) {
	deliver, unregister := hub.SetSource(sessionID)
	defer unregister()

	incoming := make(chan []byte)
	go readPtyFrames(ctx, conn, incoming)

	for {
		select {
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "")
			return
		case frame, ok := <-incoming:
			if !ok {
				return
			}
			hub.Broadcast(sessionID, frame)
		case frame, ok := <-deliver:
			if !ok {
				return
			}
			if err := conn.Write(ctx, websocket.MessageText, frame); err != nil {
				return
			}
		}
	}
}

func runPtyViewer(ctx context.Context, conn *websocket.Conn, hub *stream.PtyHub, takeover *stream.TakeoverRegistry, sessionID, userID string) {
	sub, unsubscribe := hub.Subscribe(sessionID)
	defer unsubscribe()

	incoming := make(chan []byte)
	go readPtyFrames(ctx, conn, incoming)

	for {
		select {
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "")
			return
		case frame, ok := <-incoming:
			if !ok {
				return
			}
			state, err := takeover.Get(ctx, sessionID)
			if err != nil {
				log.Printf("coop: takeover lookup failed for session %s: %v", sessionID, err)
				continue
			}
			if state.Active && state.ByID == userID {
				hub.RouteInput(sessionID, frame)
			}
		case frame, ok := <-sub:
			if !ok {
				return
			}
			if err := conn.Write(ctx, websocket.MessageText, frame); err != nil {
				return
			}
		}
	}
}

func readPtyFrames(ctx context.Context, conn *websocket.Conn, out chan<- []byte) {
	defer close(out)

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var frame ptyFrameType
		if err := json.Unmarshal(data, &frame); err != nil {
			continue
		}
		if frame.Type != "pty.output" && frame.Type != "pty.input" && frame.Type != "pty.resize" {
			continue
		}

		select {
		case out <- data:
		case <-ctx.Done():
			return
		}
	}
}
