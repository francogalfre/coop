package wsapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func ptySessionFixture(t *testing.T) (pool *db.Pool, sessionID string) {
	t.Helper()

	pool = dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "Owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	sess, err := pool.CreateAgentSession(t.Context(), "sess-pty", proj, "Owner", "/repo", "/repo", "claude-code", time.Now())
	if err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	return pool, sess.ID
}

func newPtyTestServer(pool *db.Pool, hub *stream.PtyHub, takeover *stream.TakeoverRegistry, allowedOrigins []string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/sessions/{id}/pty", withTestActor(NewPtySessionHandler(hub, takeover, pool, allowedOrigins)))

	return httptest.NewServer(mux)
}

func dialPty(ctx context.Context, wsURL, actorName string, asCLI bool) (*websocket.Conn, *http.Response, error) {
	header := http.Header{}
	header.Set(testActorHeader, actorName)
	if asCLI {
		header.Set("Authorization", "Bearer test-token")
	}

	return websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: header})
}

func readPtyFrame(t *testing.T, ctx context.Context, conn *websocket.Conn) map[string]any {
	t.Helper()

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("failed to decode frame: %v", err)
	}

	return payload
}

func TestPtySessionSourceBroadcastReachesViewer(t *testing.T) {
	pool, sessionID := ptySessionFixture(t)
	hub := stream.NewPtyHub()
	takeover := stream.NewTakeoverRegistry()

	server := newPtyTestServer(pool, hub, takeover, []string{"*"})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/" + sessionID + "/pty"

	source, _, err := dialPty(ctx, wsURL, "Owner", true)
	if err != nil {
		t.Fatalf("dial source failed: %v", err)
	}
	defer source.CloseNow()

	time.Sleep(50 * time.Millisecond)

	viewer, _, err := dialPty(ctx, wsURL, "Bob", false)
	if err != nil {
		t.Fatalf("dial viewer failed: %v", err)
	}
	defer viewer.CloseNow()

	time.Sleep(50 * time.Millisecond)

	if err := source.Write(ctx, websocket.MessageText, []byte(`{"type":"pty.output","data":"aGVsbG8="}`)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	frame := readPtyFrame(t, ctx, viewer)
	if frame["type"] != "pty.output" || frame["data"] != "aGVsbG8=" {
		t.Fatalf("unexpected frame: %+v", frame)
	}
}

func TestPtySessionViewerWithoutTakeoverInputDropped(t *testing.T) {
	pool, sessionID := ptySessionFixture(t)
	hub := stream.NewPtyHub()
	takeover := stream.NewTakeoverRegistry()

	server := newPtyTestServer(pool, hub, takeover, []string{"*"})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/" + sessionID + "/pty"

	source, _, err := dialPty(ctx, wsURL, "Owner", true)
	if err != nil {
		t.Fatalf("dial source failed: %v", err)
	}
	defer source.CloseNow()

	time.Sleep(50 * time.Millisecond)

	viewer, _, err := dialPty(ctx, wsURL, "Bob", false)
	if err != nil {
		t.Fatalf("dial viewer failed: %v", err)
	}
	defer viewer.CloseNow()

	time.Sleep(50 * time.Millisecond)

	if err := viewer.Write(ctx, websocket.MessageText, []byte(`{"type":"pty.input","data":"eA=="}`)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	shortCtx, shortCancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer shortCancel()
	if _, _, err := source.Read(shortCtx); err == nil {
		t.Fatal("source should not have received input from a viewer without takeover")
	}
}

func TestPtySessionViewerWithTakeoverInputDelivered(t *testing.T) {
	pool, sessionID := ptySessionFixture(t)
	hub := stream.NewPtyHub()
	takeover := stream.NewTakeoverRegistry()

	server := newPtyTestServer(pool, hub, takeover, []string{"*"})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/" + sessionID + "/pty"

	source, _, err := dialPty(ctx, wsURL, "Owner", true)
	if err != nil {
		t.Fatalf("dial source failed: %v", err)
	}
	defer source.CloseNow()

	time.Sleep(50 * time.Millisecond)

	viewer, _, err := dialPty(ctx, wsURL, "Bob", false)
	if err != nil {
		t.Fatalf("dial viewer failed: %v", err)
	}
	defer viewer.CloseNow()

	time.Sleep(50 * time.Millisecond)

	takeover.Set(sessionID, stream.TakeoverState{Active: true, ByID: "Bob", By: "Bob"})

	if err := viewer.Write(ctx, websocket.MessageText, []byte(`{"type":"pty.input","data":"eA=="}`)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	frame := readPtyFrame(t, ctx, source)
	if frame["type"] != "pty.input" || frame["data"] != "eA==" {
		t.Fatalf("unexpected frame: %+v", frame)
	}
}

func TestPtySessionSecondSourceReplacesFirst(t *testing.T) {
	pool, sessionID := ptySessionFixture(t)
	hub := stream.NewPtyHub()
	takeover := stream.NewTakeoverRegistry()

	server := newPtyTestServer(pool, hub, takeover, []string{"*"})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/" + sessionID + "/pty"

	first, _, err := dialPty(ctx, wsURL, "Owner", true)
	if err != nil {
		t.Fatalf("dial first source failed: %v", err)
	}
	defer first.CloseNow()

	time.Sleep(50 * time.Millisecond)

	second, _, err := dialPty(ctx, wsURL, "Owner", true)
	if err != nil {
		t.Fatalf("dial second source failed: %v", err)
	}
	defer second.CloseNow()

	time.Sleep(50 * time.Millisecond)

	viewer, _, err := dialPty(ctx, wsURL, "Bob", false)
	if err != nil {
		t.Fatalf("dial viewer failed: %v", err)
	}
	defer viewer.CloseNow()

	time.Sleep(50 * time.Millisecond)

	takeover.Set(sessionID, stream.TakeoverState{Active: true, ByID: "Bob", By: "Bob"})

	if err := viewer.Write(ctx, websocket.MessageText, []byte(`{"type":"pty.input","data":"eA=="}`)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	frame := readPtyFrame(t, ctx, second)
	if frame["type"] != "pty.input" || frame["data"] != "eA==" {
		t.Fatalf("unexpected frame on replacement source: %+v", frame)
	}

	shortCtx, shortCancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer shortCancel()
	if _, _, err := first.Read(shortCtx); err == nil {
		t.Fatal("stale first source should not have received routed input")
	}
}

func TestPtySessionRejectsUnauthenticated(t *testing.T) {
	pool, sessionID := ptySessionFixture(t)
	hub := stream.NewPtyHub()
	takeover := stream.NewTakeoverRegistry()

	server := newPtyTestServer(pool, hub, takeover, []string{"*"})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/" + sessionID + "/pty"

	_, _, err := websocket.Dial(ctx, wsURL, nil)
	if err == nil {
		t.Fatal("dial succeeded without an authenticated actor, want rejection")
	}
}

func TestPtySessionRejectsDisallowedOrigin(t *testing.T) {
	pool, sessionID := ptySessionFixture(t)
	hub := stream.NewPtyHub()
	takeover := stream.NewTakeoverRegistry()

	server := newPtyTestServer(pool, hub, takeover, []string{"http://localhost:3000"})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/" + sessionID + "/pty"

	header := http.Header{"Origin": []string{"http://evil.example"}, testActorHeader: []string{"Bob"}}
	_, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: header})
	if err == nil {
		t.Fatal("dial succeeded with a disallowed Origin, want rejection")
	}
}
