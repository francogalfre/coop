package ptywrap

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/creack/pty"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

const (
	ptyDialAttempts = 3
	ptyDialBackoff  = 500 * time.Millisecond
)

type ptyFrame struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

type ptyStreamer struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

// streamPty is a nice-to-have channel: a viewer with no reach into the pty stream must never block or crash `coop run`.
func streamPty(ctx context.Context, cfg config.Config, ptmx *os.File) (*ptyStreamer, func()) {
	s := &ptyStreamer{}
	done := make(chan struct{})

	go s.run(ctx, cfg, ptmx, done)

	return s, func() { close(done); s.closeConn() }
}

func (s *ptyStreamer) run(ctx context.Context, cfg config.Config, ptmx *os.File, done chan struct{}) {
	for attempt := 0; attempt < ptyDialAttempts; attempt++ {
		if stopped(done, ctx) {
			return
		}

		conn, err := dialPty(ctx, cfg)
		if err != nil {
			log.Printf("coop: pty stream connect: %v", err)
			if !sleepUnlessStopped(ptyDialBackoff*time.Duration(attempt+1), done, ctx) {
				return
			}
			continue
		}

		s.setConn(conn)
		if rows, cols, err := pty.Getsize(ptmx); err == nil {
			s.PushResize(cols, rows)
		}

		s.readLoop(ctx, conn, ptmx, done)
		s.setConn(nil)
	}

	log.Printf("coop: pty stream: giving up after %d attempts, continuing without remote streaming", ptyDialAttempts)
}

func (s *ptyStreamer) readLoop(ctx context.Context, conn *websocket.Conn, ptmx *os.File, done chan struct{}) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if !stopped(done, ctx) {
				log.Printf("coop: pty stream read: %v", err)
			}
			return
		}

		var frame ptyFrame
		if err := json.Unmarshal(data, &frame); err != nil {
			continue
		}

		switch frame.Type {
		case "pty.input":
			b, err := base64.StdEncoding.DecodeString(frame.Data)
			if err != nil {
				continue
			}
			if _, err := ptmx.Write(b); err != nil {
				log.Printf("coop: pty stream input write: %v", err)
			}
		case "pty.resize":
			_ = pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(frame.Rows), Cols: uint16(frame.Cols)})
		}
	}
}

// Write implements io.Writer so run.go can tee pty output into the stream via io.MultiWriter without ever failing the local copy loop.
func (s *ptyStreamer) Write(p []byte) (int, error) {
	conn := s.getConn()
	if conn == nil {
		return len(p), nil
	}

	body, err := json.Marshal(ptyFrame{Type: "pty.output", Data: base64.StdEncoding.EncodeToString(p)})
	if err != nil {
		return len(p), nil
	}

	if err := conn.Write(context.Background(), websocket.MessageText, body); err != nil {
		log.Printf("coop: pty stream write: %v", err)
	}

	return len(p), nil
}

func (s *ptyStreamer) PushResize(cols, rows int) {
	conn := s.getConn()
	if conn == nil {
		return
	}

	body, err := json.Marshal(ptyFrame{Type: "pty.resize", Cols: cols, Rows: rows})
	if err != nil {
		return
	}

	if err := conn.Write(context.Background(), websocket.MessageText, body); err != nil {
		log.Printf("coop: pty stream resize: %v", err)
	}
}

func (s *ptyStreamer) setConn(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conn = conn
}

func (s *ptyStreamer) getConn() *websocket.Conn {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn
}

func (s *ptyStreamer) closeConn() {
	conn := s.getConn()
	if conn != nil {
		_ = conn.CloseNow()
	}
}

func dialPty(ctx context.Context, cfg config.Config) (*websocket.Conn, error) {
	wsURL, err := ptyWSURL(cfg)
	if err != nil {
		return nil, err
	}

	header := http.Header{}
	if cfg.CLICredential != "" {
		header.Set("Authorization", "Bearer "+cfg.CLICredential)
	}

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		return nil, fmt.Errorf("ptywrap: dial pty stream: %w", err)
	}

	return conn, nil
}

func ptyWSURL(cfg config.Config) (string, error) {
	u, err := url.Parse(cfg.RelayURL)
	if err != nil {
		return "", fmt.Errorf("ptywrap: parse relay url: %w", err)
	}

	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	u.Path = "/v1/sessions/" + url.PathEscape(cfg.SessionID) + "/pty"

	return u.String(), nil
}

func stopped(done chan struct{}, ctx context.Context) bool {
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func sleepUnlessStopped(d time.Duration, done chan struct{}, ctx context.Context) bool {
	select {
	case <-time.After(d):
		return true
	case <-done:
		return false
	case <-ctx.Done():
		return false
	}
}
