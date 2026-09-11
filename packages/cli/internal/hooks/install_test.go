package hooks

import (
	"bytes"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

func TestInstallServesHookEventsOverHTTP(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}

	installed, err := Install(cfg, true)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	defer func() {
		if err := installed.Stop(); err != nil {
			t.Errorf("Stop() error = %v", err)
		}
	}()

	resp, err := http.Post(installed.BaseURL+"/hook/claude-code/PreToolUse", "application/json",
		bytes.NewReader([]byte(`{"hook_event_name":"PreToolUse","tool_name":"Bash"}`)))
	if err != nil {
		t.Fatalf("POST to hook server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want 200", resp.StatusCode)
	}

	relay.waitForIngested(t, 1)
}

func TestInstallFailsWhenHookAddrAlreadyBound(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve hook addr: %v", err)
	}
	defer listener.Close()

	cfg := config.Config{HookAddr: listener.Addr().String(), SessionID: "sess-a"}

	_, err = Install(cfg, true)
	if err == nil {
		t.Fatal("Install() succeeded on an already-bound address, want an explicit failure")
	}

	if !strings.Contains(err.Error(), "COOP_HOOK_ADDR") {
		t.Fatalf("error %q does not mention COOP_HOOK_ADDR", err)
	}
}
