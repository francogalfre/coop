package wsapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func newOriginGatedServer(store *stream.Store, allowedOrigins []string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/sessions/{id}/stream", NewSessionStreamHandler(store, allowedOrigins))

	return httptest.NewServer(mux)
}

func TestSessionStreamRejectsDisallowedOrigin(t *testing.T) {
	store := stream.New()
	server := newOriginGatedServer(store, []string{"http://localhost:3000"})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/sess-a/stream"

	_, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"http://evil.example"}},
	})
	if err == nil {
		t.Fatal("dial succeeded with a disallowed Origin, want rejection")
	}
}

func TestSessionStreamAcceptsAllowedOrigin(t *testing.T) {
	store := stream.New()
	server := newOriginGatedServer(store, []string{"http://localhost:3000"})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + server.URL[len("http"):] + "/v1/sessions/sess-a/stream"

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"http://localhost:3000"}},
	})
	if err != nil {
		t.Fatalf("dial failed with an allowed Origin: %v", err)
	}
	defer conn.CloseNow()

	conn.Close(websocket.StatusNormalClosure, "")
}
