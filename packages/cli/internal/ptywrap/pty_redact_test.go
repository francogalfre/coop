package ptywrap

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func TestPtyOutputRedactorRedactsSecretInSingleWrite(t *testing.T) {
	sink := &syncBuffer{}
	w := newPtyOutputRedactor(sink)

	if _, err := w.Write([]byte("token=sk-abcdefghijklmnopqrstuv\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got := sink.String()
	if strings.Contains(got, "sk-abcdefghijklmnopqrstuv") {
		t.Fatalf("secret leaked to sink: %q", got)
	}
	if !strings.Contains(got, "[redacted]") {
		t.Fatalf("sink missing redaction marker: %q", got)
	}
}

func TestPtyOutputRedactorCatchesSecretSplitAcrossWrites(t *testing.T) {
	sink := &syncBuffer{}
	w := newPtyOutputRedactor(sink)

	secret := "sk-abcdefghijklmnopqrstuv"
	split := len("token=sk-abcdefghij")
	full := "token=" + secret + "\n"

	if _, err := w.Write([]byte(full[:split])); err != nil {
		t.Fatalf("Write 1: %v", err)
	}
	if strings.Contains(sink.String(), "sk-") {
		t.Fatalf("partial secret leaked before line completed: %q", sink.String())
	}

	if _, err := w.Write([]byte(full[split:])); err != nil {
		t.Fatalf("Write 2: %v", err)
	}

	got := sink.String()
	if strings.Contains(got, secret) {
		t.Fatalf("secret split across Write() calls leaked: %q", got)
	}
	if !strings.Contains(got, "[redacted]") {
		t.Fatalf("sink missing redaction marker: %q", got)
	}
}

func TestPtyOutputRedactorPassesThroughNormalOutput(t *testing.T) {
	sink := &syncBuffer{}
	w := newPtyOutputRedactor(sink)

	lines := []string{"building project\n", "3 files changed\n", "done.\n"}
	for _, line := range lines {
		if _, err := w.Write([]byte(line)); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	want := strings.Join(lines, "")
	if got := sink.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPtyOutputRedactorFlushesOversizedUnterminatedLine(t *testing.T) {
	sink := &syncBuffer{}
	w := newPtyOutputRedactor(sink)

	// no newline anywhere: simulates a \r-driven progress bar with no line terminator.
	chunk := bytes.Repeat([]byte("."), redactMaxBuffer+1)

	if _, err := w.Write(chunk); err != nil {
		t.Fatalf("Write: %v", err)
	}

	w.mu.Lock()
	pending := len(w.pending)
	w.mu.Unlock()

	if pending != 0 {
		t.Fatalf("pending buffer not flushed after exceeding max size: %d bytes still held", pending)
	}
	if sink.String() != string(chunk) {
		t.Fatalf("sink content corrupted for oversized flush")
	}
}

func TestPtyOutputRedactorIdleFlushesTrailingPartialLine(t *testing.T) {
	sink := &syncBuffer{}
	w := newPtyOutputRedactor(sink)

	if _, err := w.Write([]byte("Enter password: ")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if sink.String() != "" {
		t.Fatalf("partial line flushed before idle timeout: %q", sink.String())
	}

	time.Sleep(redactIdleFlush + 100*time.Millisecond)

	if got := sink.String(); got != "Enter password: " {
		t.Fatalf("got %q after idle flush, want %q", got, "Enter password: ")
	}
}

func TestPtyOutputRedactorCloseFlushesTrailingPartialLine(t *testing.T) {
	sink := &syncBuffer{}
	w := newPtyOutputRedactor(sink)

	if _, err := w.Write([]byte("no trailing newline")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	w.Close()

	if got := sink.String(); got != "no trailing newline" {
		t.Fatalf("got %q after Close, want %q", got, "no trailing newline")
	}
}

func TestPtyOutputRedactorThroughPtyStreamerFramesRoundTrip(t *testing.T) {
	connCh := make(chan fakePtyConn, 1)
	server := startFakePtyServer(t, connCh)
	defer server.Close()

	ptmx, _ := openRawPty(t)

	cfg := config.Config{RelayURL: server.URL, SessionID: "sess-redact"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamer, stop := streamPty(ctx, cfg, ptmx)
	defer stop()

	fc := acceptFakePtyConn(t, connCh)
	defer fc.conn.CloseNow()

	readFrame(t, fc.conn) // initial resize-on-connect frame

	w := newPtyOutputRedactor(streamer)

	secret := "sk-abcdefghijklmnopqrstuv"
	line := "leaked=" + secret + "\n"
	half := len(line) / 2

	if _, err := w.Write([]byte(line[:half])); err != nil {
		t.Fatalf("Write 1: %v", err)
	}
	if _, err := w.Write([]byte(line[half:])); err != nil {
		t.Fatalf("Write 2: %v", err)
	}

	frame := readFrame(t, fc.conn)
	if frame.Type != "pty.output" {
		t.Fatalf("got type %q, want pty.output", frame.Type)
	}

	data, err := base64.StdEncoding.DecodeString(frame.Data)
	if err != nil {
		t.Fatalf("decode base64: %v", err)
	}

	if strings.Contains(string(data), secret) {
		t.Fatalf("secret reached outbound frame: %q", data)
	}
	if !strings.Contains(string(data), "[redacted]") {
		t.Fatalf("outbound frame missing redaction marker: %q", data)
	}
}
