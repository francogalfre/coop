package projectcontext

import (
	"context"
	"log"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/relayclient"
)

const deliverTimeout = 10 * time.Second

// Sharing context is a nicety, never a reason to block session startup - every failure here is logged and swallowed.
func Deliver(ctx context.Context, cfg config.Config) {
	if cfg.Project == "" {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deliverTimeout)
	defer cancel()

	projectCtx, err := relayclient.GetProjectContext(ctx, cfg, cfg.Project)
	if err != nil {
		log.Printf("coop: get project context: %v", err)
		return
	}

	if projectCtx.Text == "" {
		return
	}

	version := projectCtx.Version
	if err := relayclient.DeliverSteer(ctx, cfg, cfg.SessionID, projectCtx.Text, &version); err != nil {
		log.Printf("coop: deliver project context: %v", err)
	}
}
