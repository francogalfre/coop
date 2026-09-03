package ptywrap

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

func TestDeliverPendingSteerWritesCommandWithoutAttribution(t *testing.T) {
	relay := &fakeSteerRelay{}
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setReceipt("Alice", "/model sonnet", "steer-1", "command")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	deliverPendingSteer(ctx, cfg, w, false)
	_ = w.Close()

	out, _ := io.ReadAll(r)

	want := "/model sonnet\r"
	if string(out) != want {
		t.Fatalf("got %q, want %q (a command is a keystroke, never wrapped in attribution)", out, want)
	}
}

func TestDeliverPendingSteerDropsCommandNotOnAllowlist(t *testing.T) {
	relay := &fakeSteerRelay{}
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setReceipt("Alice", "/rm -rf /", "steer-1", "command")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	deliverPendingSteer(ctx, cfg, w, false)
	_ = w.Close()

	out, _ := io.ReadAll(r)
	if len(out) != 0 {
		t.Fatalf("got %q written to the pty, want nothing for a command off the allowlist", out)
	}
}

func TestDeliverPendingSteerEmitsSteerDeliveredForCommand(t *testing.T) {
	relay := &fakeSteerRelay{}
	relayServer := relay.server()
	defer relayServer.Close()
	relay.setReceipt("Alice", "/status", "steer-2", "command")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	deliverPendingSteer(ctx, cfg, w, false)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, e := range relay.eventsSnapshot() {
			if e.body["type"] == "steer.delivered" {
				if e.body["steer_id"] != "steer-2" {
					t.Fatalf("got steer_id %v, want steer-2", e.body["steer_id"])
				}
				if _, ok := e.body["hook_event"]; ok {
					t.Fatalf("got hook_event %v, want it omitted in run mode", e.body["hook_event"])
				}
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("no steer.delivered event posted")
}

func TestDeliverPendingSteerSkipsSteerDeliveredWhenIDAbsent(t *testing.T) {
	relay := &fakeSteerRelay{}
	relayServer := relay.server()
	defer relayServer.Close()
	relay.set("Alice", "try the other branch")

	cfg := config.Config{RelayURL: relayServer.URL, SessionID: "sess-a"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	deliverPendingSteer(ctx, cfg, w, false)

	time.Sleep(100 * time.Millisecond)

	for _, e := range relay.eventsSnapshot() {
		if e.body["type"] == "steer.delivered" {
			t.Fatalf("got a steer.delivered event with no id in the relay response: %v", e.body)
		}
	}
}
