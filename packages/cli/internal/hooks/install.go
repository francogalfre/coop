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
// Trying cfg.HookAddr first means a restarted `coop attach` usually reuses the
// same URL, so stale harness config entries self-heal after a crash.
func Install(cfg config.Config, steer bool) (*Installed, error) {
	listener, err := listenHook(cfg.HookAddr, cfg.HookAddrExplicit)
	if err != nil {
		return nil, err
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

func listenHook(addr string, explicit bool) (net.Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		return listener, nil
	}

	if explicit {
		return nil, fmt.Errorf("hooks: listen on %s: %w; another coop session may be holding this address, set COOP_HOOK_ADDR to a free one", addr, err)
	}

	// Falling back is only safe because callers wire harnesses to the address we
	// actually bound, never to the one we asked for.
	fallback, fallbackErr := net.Listen("tcp", freePortOn(addr))
	if fallbackErr != nil {
		return nil, fmt.Errorf("hooks: listen on %s: %w; retrying on a free port also failed: %w", addr, err, fallbackErr)
	}

	return fallback, nil
}

func freePortOn(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "127.0.0.1:0"
	}

	return net.JoinHostPort(host, "0")
}
