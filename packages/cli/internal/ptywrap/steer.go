package ptywrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/commands"
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
		deliverMailboxMessage(cfg, ptmx, steer)
	}

	if !steer.Takeover.Active {
		return false
	}

	if !notified {
		notice := fmt.Sprintf("\r[coop] %s has taken over this session\r", steer.Takeover.By)
		if _, err := ptmx.Write([]byte(notice)); err != nil {
			log.Printf("coop: steer write: %v", err)
		}
	}

	return true
}

func deliverMailboxMessage(cfg config.Config, ptmx *os.File, steer relayclient.SteerResult) {
	if steer.Kind == "command" {
		deliverCommand(cfg, ptmx, steer)
		return
	}

	msg := fmt.Sprintf("\r[%s via coop] %s\r", steer.From, steer.Text)
	if _, err := ptmx.Write([]byte(msg)); err != nil {
		log.Printf("coop: steer write: %v", err)
		return
	}

	if steer.ID != "" {
		postSteerDelivered(cfg, steer.ID)
	}
}

// deliverCommand writes the command exactly as a human would type it - no
// "[name via coop]" attribution, since a command is a keystroke sequence,
// not attributed context (security.md). It re-validates against the
// allowlist itself rather than trusting the relay: the CLI is the last line
// before a real keystroke reaches the harness's terminal.
func deliverCommand(cfg config.Config, ptmx *os.File, steer relayclient.SteerResult) {
	if !commands.Validate(steer.Text) {
		log.Printf("coop: dropping harness command %q: not on the allowlist", steer.Text)
		return
	}

	if _, err := ptmx.Write([]byte(steer.Text + "\r")); err != nil {
		log.Printf("coop: command write: %v", err)
		return
	}

	if steer.ID != "" {
		postSteerDelivered(cfg, steer.ID)
	}
}

func postSteerDelivered(cfg config.Config, steerID string) {
	postSessionEvent(cfg, nextSelfSeq(), map[string]any{"type": "steer.delivered", "steer_id": steerID})
}
