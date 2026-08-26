package wsapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func newTestServer(store *stream.Store) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/sessions/{id}/stream", NewSessionStreamHandler(store, stream.NewPresenceHub(), []string{"*"}))

	return httptest.NewServer(mux)
}

func readEvent(t *testing.T, ctx context.Context, conn *websocket.Conn) map[string]any {
	t.Helper()

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}

	return payload
}

func TestSessionStreamBackfillThenLiveTail(t *testing.T) {
	store := stream.New()
	store.Append("sess-a", json.RawMessage(`{"n":1}`))

	server := newTestServer(store)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/sess-a/stream"

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.CloseNow()

	backfill := readEvent(t, ctx, conn)
	if backfill["n"] != float64(1) {
		t.Fatalf("unexpected backfill event: %+v", backfill)
	}

	store.Append("sess-a", json.RawMessage(`{"n":2}`))

	live := readEvent(t, ctx, conn)
	if live["n"] != float64(2) {
		t.Fatalf("unexpected live event: %+v", live)
	}

	conn.Close(websocket.StatusNormalClosure, "")
}

func TestSessionStreamPresenceTypingBroadcastToOtherOnly(t *testing.T) {
	store := stream.New()

	server := newTestServer(store)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/sess-a/stream"

	sender, _, err := websocket.Dial(ctx, wsURL+"?name=Alice", nil)
	if err != nil {
		t.Fatalf("dial sender failed: %v", err)
	}
	defer sender.CloseNow()

	time.Sleep(50 * time.Millisecond)

	other, _, err := websocket.Dial(ctx, wsURL+"?name=Bob", nil)
	if err != nil {
		t.Fatalf("dial other failed: %v", err)
	}
	defer other.CloseNow()

	join := readEvent(t, ctx, sender)
	if join["type"] != "human.join" {
		t.Fatalf("expected human.join, got %+v", join)
	}

	if err := sender.Write(ctx, websocket.MessageText, []byte(`{"type":"presence.typing","active":true}`)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	typing := readEvent(t, ctx, other)
	if typing["kind"] != "presence" || typing["type"] != "presence.typing" {
		t.Fatalf("unexpected message: %+v", typing)
	}
	actor, _ := typing["actor"].(map[string]any)
	if actor["name"] != "Alice" {
		t.Fatalf("unexpected actor: %+v", actor)
	}
	if typing["active"] != true {
		t.Fatalf("expected active true, got %+v", typing["active"])
	}

	shortCtx, shortCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer shortCancel()
	if _, _, err := sender.Read(shortCtx); err == nil {
		t.Fatal("sender should not have received its own typing broadcast")
	}
}

func TestSessionStreamHumanJoinAndLeaveBroadcast(t *testing.T) {
	store := stream.New()

	server := newTestServer(store)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/sess-a/stream"

	watcher, _, err := websocket.Dial(ctx, wsURL+"?name=Alice", nil)
	if err != nil {
		t.Fatalf("dial watcher failed: %v", err)
	}
	defer watcher.CloseNow()

	time.Sleep(50 * time.Millisecond)

	joiner, _, err := websocket.Dial(ctx, wsURL+"?name=Bob", nil)
	if err != nil {
		t.Fatalf("dial joiner failed: %v", err)
	}

	join := readEvent(t, ctx, watcher)
	if join["type"] != "human.join" {
		t.Fatalf("expected human.join, got %+v", join)
	}
	actor, _ := join["actor"].(map[string]any)
	if actor["name"] != "Bob" {
		t.Fatalf("unexpected actor: %+v", actor)
	}

	joiner.Close(websocket.StatusNormalClosure, "")

	leave := readEvent(t, ctx, watcher)
	if leave["type"] != "human.leave" {
		t.Fatalf("expected human.leave, got %+v", leave)
	}
	actor, _ = leave["actor"].(map[string]any)
	if actor["name"] != "Bob" {
		t.Fatalf("unexpected actor: %+v", actor)
	}
}
