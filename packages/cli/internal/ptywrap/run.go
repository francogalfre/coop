package ptywrap

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"golang.org/x/term"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

func Run(ctx context.Context, cfg config.Config, harnessName, name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("ptywrap: start %s: %w", name, err)
	}
	defer ptmx.Close()

	postSessionStart(cfg, harnessName)
	defer postSessionEnd(cfg)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("ptywrap: put terminal in raw mode: %w", err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	streamer, stopStream := streamPty(ctx, cfg, ptmx)
	defer stopStream()

	stopResize := watchResize(ptmx, streamer.PushResize)
	defer stopResize()

	stopSteer := watchSteer(ctx, cfg, ptmx)
	defer stopSteer()

	outputDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.MultiWriter(os.Stdout, streamer), ptmx)
		close(outputDone)
	}()

	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	waitErr := cmd.Wait()
	<-outputDone

	if waitErr != nil {
		return fmt.Errorf("ptywrap: %s: %w", name, waitErr)
	}

	return nil
}
