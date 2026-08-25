package ptywrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/relayclient"
)

const steerPollInterval = 1500 * time.Millisecond

func watchSteer(ctx context.Context, cfg config.Config, ptmx *os.File) func() {
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(steerPollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-ticker.C:
				deliverPendingSteer(ctx, cfg, ptmx)
			}
		}
	}()

	return func() { close(done) }
}

func deliverPendingSteer(ctx context.Context, cfg config.Config, ptmx *os.File) {
	from, text, ok, err := relayclient.GetSteer(ctx, cfg, cfg.SessionID)
	if err != nil {
		log.Printf("coop: steer poll: %v", err)
		return
	}

	if !ok {
		return
	}

	msg := fmt.Sprintf("\n[%s via coop] %s\n", from, text)
	if _, err := ptmx.Write([]byte(msg)); err != nil {
		log.Printf("coop: steer write: %v", err)
	}
}
