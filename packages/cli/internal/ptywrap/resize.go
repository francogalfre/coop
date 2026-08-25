package ptywrap

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
)

func watchResize(ptmx *os.File) func() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	resize := func() {
		_ = pty.InheritSize(os.Stdin, ptmx)
	}
	resize()

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-sig:
				resize()
			case <-done:
				return
			}
		}
	}()

	return func() {
		signal.Stop(sig)
		close(done)
	}
}
