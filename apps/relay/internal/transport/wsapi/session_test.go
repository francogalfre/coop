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
	mux.HandleFunc("GET /v1/sessions/{id}/stream", NewSessionStreamHandler(store))

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
