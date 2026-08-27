package wsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const testActorHeader = "X-Test-Actor-Name"

func withTestActor(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.Header.Get(testActorHeader)
		if name == "" {
			http.NotFound(w, r)
			return
		}

		next(w, r.WithContext(auth.WithActor(r.Context(), auth.Actor{UserID: name, DisplayName: name})))
	}
}

func dialAs(ctx context.Context, wsURL, name string) (*websocket.Conn, *http.Response, error) {
	header := http.Header{}
	header.Set(testActorHeader, name)

	return websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: header})
}

func newTestServer(store *stream.Store) *httptest.Server {
	return newTestServerWithPool(nil, store)
}

func newTestServerWithPool(pool *db.Pool, store *stream.Store) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/sessions/{id}/stream", withTestActor(NewSessionStreamHandler(pool, store, stream.NewPresenceHub(), []string{"*"})))

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

	conn, _, err := dialAs(ctx, wsURL, "Tester")
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

	sender, _, err := dialAs(ctx, wsURL, "Alice")
	if err != nil {
		t.Fatalf("dial sender failed: %v", err)
	}
	defer sender.CloseNow()

	time.Sleep(50 * time.Millisecond)

	other, _, err := dialAs(ctx, wsURL, "Bob")
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

	watcher, _, err := dialAs(ctx, wsURL, "Alice")
	if err != nil {
		t.Fatalf("dial watcher failed: %v", err)
	}
	defer watcher.CloseNow()

	time.Sleep(50 * time.Millisecond)

	joiner, _, err := dialAs(ctx, wsURL, "Bob")
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

func TestSessionStreamClosesConnectionOnOversizedFrame(t *testing.T) {
	store := stream.New()

	server := newTestServer(store)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/sess-a/stream"

	conn, _, err := dialAs(ctx, wsURL, "Attacker")
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.CloseNow()

	oversized := append([]byte(`{"type":"presence.typing","active":true,"pad":"`), bytes.Repeat([]byte("a"), maxSessionFrameBytes+1)...)
	oversized = append(oversized, []byte(`"}`)...)

	if err := conn.Write(ctx, websocket.MessageText, oversized); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	shortCtx, shortCancel := context.WithTimeout(ctx, time.Second)
	defer shortCancel()
	if _, _, err := conn.Read(shortCtx); err == nil {
		t.Fatal("expected the connection to be closed after an oversized frame")
	}
}

func TestSessionStreamBackfillsFromPostgresAfterSimulatedRestart(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if _, err := pool.CreateAgentSession(t.Context(), "sess-restart", proj, "user-owner", "/repo", "/repo", "claude-code", time.Now()); err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	if _, err := pool.AppendEvent(t.Context(), "sess-restart", []byte(`{"type":"session.start","harness":"claude-code"}`)); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
	if _, err := pool.AppendEvent(t.Context(), "sess-restart", []byte(`{"type":"file.touched","path":"src/foo.ts"}`)); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}

	// A fresh, empty store simulates a relay restart: only Postgres remains.
	freshStore := stream.New()
	server := newTestServerWithPool(pool, freshStore)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/sess-restart/stream"

	conn, _, err := dialAs(ctx, wsURL, "Tester")
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.CloseNow()

	first := readEvent(t, ctx, conn)
	if first["type"] != "session.start" {
		t.Fatalf("expected session.start to survive the simulated restart, got %+v", first)
	}
	if first["seq"] != float64(1) {
		t.Fatalf("got seq %v, want 1", first["seq"])
	}

	second := readEvent(t, ctx, conn)
	if second["type"] != "file.touched" {
		t.Fatalf("expected file.touched next, got %+v", second)
	}
	if second["seq"] != float64(2) {
		t.Fatalf("got seq %v, want 2", second["seq"])
	}

	conn.Close(websocket.StatusNormalClosure, "")
}
