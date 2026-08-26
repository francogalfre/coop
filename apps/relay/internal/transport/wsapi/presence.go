package wsapi

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

type presenceActor struct {
	Name string `json:"name"`
}

type presenceFrame struct {
	Kind   string        `json:"kind"`
	Type   string        `json:"type"`
	Actor  presenceActor `json:"actor"`
	Active *bool         `json:"active,omitempty"`
}

type clientPresenceMessage struct {
	Type   string `json:"type"`
	Active bool   `json:"active"`
}

func readClientFrames(ctx context.Context, conn *websocket.Conn, out chan<- clientPresenceMessage) {
	defer close(out)

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var msg clientPresenceMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		if msg.Type != "presence.typing" {
			continue
		}

		select {
		case out <- msg:
		case <-ctx.Done():
			return
		}
	}
}

func broadcastPresence(hub *stream.PresenceHub, sessionID string, except chan []byte, name, eventType string) {
	frame := presenceFrame{Kind: "presence", Type: eventType, Actor: presenceActor{Name: name}}

	data, err := json.Marshal(frame)
	if err != nil {
		return
	}

	hub.Broadcast(sessionID, except, data)
}

func broadcastTyping(hub *stream.PresenceHub, sessionID string, except chan []byte, name string, active bool) {
	frame := presenceFrame{Kind: "presence", Type: "presence.typing", Actor: presenceActor{Name: name}, Active: &active}

	data, err := json.Marshal(frame)
	if err != nil {
		return
	}

	hub.Broadcast(sessionID, except, data)
}
