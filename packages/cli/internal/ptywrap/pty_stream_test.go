package ptywrap

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/creack/pty"
	"golang.org/x/term"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

type fakePtyConn struct {
	conn *websocket.Conn
	auth string
}

func startFakePtyServer(t *testing.T, connCh chan<- fakePtyConn) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/sessions/{id}/pty", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		connCh <- fakePtyConn{conn: conn, auth: r.Header.Get("Authorization")}
	})

	return httptest.NewServer(mux)
}

// Raw mode disables the tty's line discipline so bytes written on the master arrive on the slave immediately, unbuffered and unechoed.
func openRawPty(t *testing.T) (*os.File, *os.File) {
	t.Helper()

	ptmx, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open: %v", err)
	}
	t.Cleanup(func() {
		_ = ptmx.Close()
		_ = tty.Close()
	})

	if _, err := term.MakeRaw(int(tty.Fd())); err != nil {
		t.Fatalf("term.MakeRaw: %v", err)
	}

	return ptmx, tty
}

func acceptFakePtyConn(t *testing.T, connCh <-chan fakePtyConn) fakePtyConn {
	t.Helper()

	select {
	case fc := <-connCh:
		return fc
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for pty stream connection")
		return fakePtyConn{}
	}
}

func readFrame(t *testing.T, conn *websocket.Conn) ptyFrame {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}

	var frame ptyFrame
	if err := json.Unmarshal(data, &frame); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}

	return frame
}

func TestStreamPtySendsOutputAsBase64Frames(t *testing.T) {
	connCh := make(chan fakePtyConn, 1)
	server := startFakePtyServer(t, connCh)
	defer server.Close()

	ptmx, _ := openRawPty(t)

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-a", CLICredential: "tok-a"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamer, stop := streamPty(ctx, cfg, ptmx)
	defer stop()

	fc := acceptFakePtyConn(t, connCh)
	defer fc.conn.CloseNow()

	if fc.auth != "Bearer tok-a" {
		t.Errorf("got Authorization %q, want Bearer tok-a", fc.auth)
	}

	readFrame(t, fc.conn) // initial resize-on-connect frame

	if _, err := streamer.Write([]byte("hello")); err != nil {
		t.Fatalf("streamer.Write: %v", err)
	}

	frame := readFrame(t, fc.conn)
	if frame.Type != "pty.output" {
		t.Fatalf("got type %q, want pty.output", frame.Type)
	}

	data, err := base64.StdEncoding.DecodeString(frame.Data)
	if err != nil {
		t.Fatalf("decode base64: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("got data %q, want %q", data, "hello")
	}
}

func TestStreamPtyAppliesIncomingInputToPty(t *testing.T) {
	connCh := make(chan fakePtyConn, 1)
	server := startFakePtyServer(t, connCh)
	defer server.Close()

	ptmx, tty := openRawPty(t)

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-b"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, stop := streamPty(ctx, cfg, ptmx)
	defer stop()

	fc := acceptFakePtyConn(t, connCh)
	defer fc.conn.CloseNow()

	readFrame(t, fc.conn) // initial resize-on-connect frame

	body, err := json.Marshal(ptyFrame{Type: "pty.input", Data: base64.StdEncoding.EncodeToString([]byte("ls\n"))})
	if err != nil {
		t.Fatalf("marshal frame: %v", err)
	}
	if err := fc.conn.Write(ctx, websocket.MessageText, body); err != nil {
		t.Fatalf("write frame: %v", err)
	}

	readDone := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 3)
		n, _ := io.ReadFull(tty, buf)
		readDone <- buf[:n]
	}()

	select {
	case got := <-readDone:
		if string(got) != "ls\n" {
			t.Errorf("got %q, want %q", got, "ls\n")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out reading input from tty")
	}
}

func TestStreamPtySendsResizeFrames(t *testing.T) {
	connCh := make(chan fakePtyConn, 1)
	server := startFakePtyServer(t, connCh)
	defer server.Close()

	ptmx, _ := openRawPty(t)
	if err := pty.Setsize(ptmx, &pty.Winsize{Rows: 40, Cols: 100}); err != nil {
		t.Fatalf("pty.Setsize: %v", err)
	}

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-c"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamer, stop := streamPty(ctx, cfg, ptmx)
	defer stop()

	fc := acceptFakePtyConn(t, connCh)
	defer fc.conn.CloseNow()

	initial := readFrame(t, fc.conn)
	if initial.Type != "pty.resize" || initial.Cols != 100 || initial.Rows != 40 {
		t.Fatalf("got initial resize %+v, want {type:pty.resize cols:100 rows:40}", initial)
	}

	streamer.PushResize(120, 50)

	next := readFrame(t, fc.conn)
	if next.Type != "pty.resize" || next.Cols != 120 || next.Rows != 50 {
		t.Fatalf("got resize %+v, want {type:pty.resize cols:120 rows:50}", next)
	}
}
