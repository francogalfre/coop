package ptywrap

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/user"
	"sync/atomic"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/capabilities"
	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/projectcontext"
	"github.com/francogalfre/coop/packages/cli/internal/relayclient"
	"github.com/francogalfre/coop/packages/cli/internal/repoid"
)

const postEventTimeout = 15 * time.Second

var selfEventSeq atomic.Int64

// nextSelfSeq numbers events ptywrap posts about itself (session start/end,
// steer.delivered) - a separate counter from hooks.Server's, since run mode
// runs both an ingest server (for the wrapped harness's own hook events) and
// this self-event stream side by side.
func nextSelfSeq() int {
	return int(selfEventSeq.Add(1)) - 1
}

func postSessionStart(cfg config.Config, harnessName string) {
	cwd, _ := os.Getwd()

	fields := map[string]any{
		"type":         "session.start",
		"harness":      harnessName,
		"cwd":          cwd,
		"owner":        ownerFields(cfg),
		"capabilities": capabilities.ForRun(),
	}

	if repo := repoid.Detect(cwd); repo != "" {
		fields["repo"] = repo
	}

	postSessionEvent(cfg, nextSelfSeq(), fields)

	projectcontext.Deliver(context.Background(), cfg)
}

func postSessionEnd(cfg config.Config) {
	postSessionEvent(cfg, nextSelfSeq(), map[string]any{"type": "session.end"})
}

func postSessionEvent(cfg config.Config, seq int, fields map[string]any) {
	envelope := map[string]any{
		"v":          1,
		"session_id": cfg.SessionID,
		"seq":        seq,
		"ts":         time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range fields {
		envelope[k] = v
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		log.Printf("coop: marshal %v event: %v", fields["type"], err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), postEventTimeout)
	defer cancel()

	if err := relayclient.PostEvent(ctx, cfg, body); err != nil {
		log.Printf("coop: post %v event: %v", fields["type"], err)
	}
}

func ownerFields(cfg config.Config) map[string]string {
	if cfg.Username != "" || cfg.DisplayName != "" {
		id, name := cfg.Username, cfg.DisplayName
		if id == "" {
			id = name
		}
		if name == "" {
			name = id
		}
		return map[string]string{"id": id, "display_name": name}
	}

	id, name := "local", "local"
	if u, err := user.Current(); err == nil && u.Username != "" {
		id, name = u.Username, u.Username
	}

	return map[string]string{"id": id, "display_name": name}
}
