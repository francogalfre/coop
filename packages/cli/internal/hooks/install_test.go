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

func TestInstallFailsWhenAnExplicitHookAddrIsTaken(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve hook addr: %v", err)
	}
	defer listener.Close()

	cfg := config.Config{
		HookAddr:         listener.Addr().String(),
		HookAddrExplicit: true,
		SessionID:        "sess-a",
	}

	_, err = Install(cfg, true)
	if err == nil {
		t.Fatal("Install() succeeded on an already-bound explicit address, want an explicit failure")
	}

	if !strings.Contains(err.Error(), "COOP_HOOK_ADDR") {
		t.Fatalf("error %q does not mention COOP_HOOK_ADDR", err)
	}
}

func TestInstallPicksAFreePortWhenTheDefaultIsTaken(t *testing.T) {
	relay := newFakeRelay()
	relayServer := relay.server()
	defer relayServer.Close()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve hook addr: %v", err)
	}
	defer listener.Close()

	taken := listener.Addr().String()
	cfg := config.Config{
		RelayURL:  relayServer.URL,
		HookAddr:  taken,
		SessionID: "sess-a",
	}

	installed, err := Install(cfg, true)
	if err != nil {
		t.Fatalf("Install() error = %v, want a fallback to a free port", err)
	}
	defer func() { _ = installed.Stop() }()

	if strings.Contains(installed.BaseURL, taken) {
		t.Fatalf("BaseURL %q reuses the taken address %q", installed.BaseURL, taken)
	}

	resp, err := http.Post(installed.BaseURL+"/hook/claude-code/PreToolUse", "application/json",
		bytes.NewReader([]byte(`{"hook_event_name":"PreToolUse","tool_name":"Bash"}`)))
	if err != nil {
		t.Fatalf("POST to the fallback hook server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d from the fallback server, want 200", resp.StatusCode)
	}
}
