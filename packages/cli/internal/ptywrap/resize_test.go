package ptywrap

import (
	"testing"
	"time"

	"github.com/creack/pty"
)

func TestWatchResizeInvokesCallbackWithCurrentSize(t *testing.T) {
	ptmx, _ := openRawPty(t)

	if err := pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		t.Fatalf("pty.Setsize: %v", err)
	}

	calls := make(chan [2]int, 1)
	stop := watchResize(ptmx, func(cols, rows int) {
		calls <- [2]int{cols, rows}
	})
	defer stop()

	select {
	case got := <-calls:
		if got != [2]int{80, 24} {
			t.Errorf("got (cols,rows)=%v, want (80,24)", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for onResize callback")
	}
}
