package hooks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

const shutdownTimeout = 5 * time.Second

type Installed struct {
	BaseURL string
	Stop    func() error
}

// Install starts the hook ingest server only. Wiring an actual harness up
// to it (writing settings.local.json, a plugin file, an extension file) is
// internal/harness's job, once it has BaseURL.
//
// It binds cfg.HookAddr first so a restarted `coop attach` usually reuses
// the same URL and stale harness config entries self-heal after a crash.
func Install(cfg config.Config, steer bool) (*Installed, error) {
	listener, err := listenHook(cfg.HookAddr)
	if err != nil {
		return nil, fmt.Errorf("hooks: listen: %w", err)
	}

	baseURL := fmt.Sprintf("http://%s", listener.Addr().String())

	srv := NewServer(cfg, steer)
	httpServer := &http.Server{Handler: srv.Handler()}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpServer.Serve(listener)
	}()

	stop := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("hooks: shutdown: %w", err)
		}

		if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("hooks: serve: %w", err)
		}

		srv.Close()

		return nil
	}

	return &Installed{BaseURL: baseURL, Stop: stop}, nil
}

func listenHook(addr string) (net.Listener, error) {
	if listener, err := net.Listen("tcp", addr); err == nil {
		return listener, nil
	}

	return net.Listen("tcp", "127.0.0.1:0")
}
