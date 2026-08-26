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
	notified := false

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
				notified = deliverPendingSteer(ctx, cfg, ptmx, notified)
			}
		}
	}()

	return func() { close(done) }
}

// A pty has no veto primitive, so a takeover notice here is advisory only; notified edge-triggers it once per claim instead of every poll.
func deliverPendingSteer(ctx context.Context, cfg config.Config, ptmx *os.File, notified bool) bool {
	steer, err := relayclient.GetSteer(ctx, cfg, cfg.SessionID)
	if err != nil {
		log.Printf("coop: steer poll: %v", err)
		return notified
	}

	if steer.HasMessage {
		msg := fmt.Sprintf("\n[%s via coop] %s\n", steer.From, steer.Text)
		if _, err := ptmx.Write([]byte(msg)); err != nil {
			log.Printf("coop: steer write: %v", err)
		}
	}

	if !steer.Takeover.Active {
		return false
	}

	if !notified {
		notice := fmt.Sprintf("\n[coop] %s has taken over this session\n", steer.Takeover.By)
		if _, err := ptmx.Write([]byte(notice)); err != nil {
			log.Printf("coop: steer write: %v", err)
		}
	}

	return true
}
