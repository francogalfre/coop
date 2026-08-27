package ptywrap

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

const (
	redactMaxBuffer = 64 * 1024
	redactIdleFlush = 200 * time.Millisecond
)

// ptyOutputRedactor line-buffers pty output before it reaches sink so a secret split across two Write() calls by the pty's arbitrary chunking is still whole when redact.Redactor sees it.
type ptyOutputRedactor struct {
	mu      sync.Mutex
	sink    io.Writer
	red     *redact.Redactor
	pending []byte
	timer   *time.Timer
}

func newPtyOutputRedactor(sink io.Writer) *ptyOutputRedactor {
	return &ptyOutputRedactor{sink: sink, red: redact.New()}
}

func (w *ptyOutputRedactor) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.pending = append(w.pending, p...)
	w.flushComplete()

	if len(w.pending) > 0 {
		w.armIdleFlush()
	}

	return len(p), nil
}

// flushComplete emits every full line in pending and leaves a trailing partial line buffered until more data completes it, the idle timer fires, or it grows past redactMaxBuffer (a progress bar using \r with no newline).
func (w *ptyOutputRedactor) flushComplete() {
	for {
		i := bytes.IndexByte(w.pending, '\n')
		if i < 0 {
			break
		}
		w.emit(w.pending[:i+1])
		w.pending = w.pending[i+1:]
	}

	if len(w.pending) > redactMaxBuffer {
		w.emit(w.pending)
		w.pending = nil
	}
}

// emit must be called with mu held; sink write errors never surface, matching ptyStreamer.Write's never-fail contract.
func (w *ptyOutputRedactor) emit(chunk []byte) {
	if len(chunk) == 0 {
		return
	}

	redacted, _ := w.red.Text(string(chunk))
	_, _ = w.sink.Write([]byte(redacted))
}

func (w *ptyOutputRedactor) armIdleFlush() {
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(redactIdleFlush, w.flushIdle)
}

func (w *ptyOutputRedactor) flushIdle() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.emit(w.pending)
	w.pending = nil
}

// Close flushes any buffered partial line, e.g. an unterminated prompt, once the pty has nothing more to write.
func (w *ptyOutputRedactor) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.timer != nil {
		w.timer.Stop()
	}

	w.emit(w.pending)
	w.pending = nil
}
