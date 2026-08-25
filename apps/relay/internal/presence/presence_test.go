package presence

import (
	"testing"
	"time"
)

func TestSessionLifecycle(t *testing.T) {
	r := New()
	start := time.Now()

	r.SessionStarted("sess-a", "/repo", "Alice", "claude-code", start)

	active := r.ActiveSessions("/repo")
	if len(active) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(active))
	}
	if active[0].SessionID != "sess-a" || active[0].Owner != "Alice" || active[0].Harness != "claude-code" {
		t.Fatalf("unexpected session summary: %+v", active[0])
	}

	r.SessionEnded("sess-a", start.Add(time.Minute))

	active = r.ActiveSessions("/repo")
	if len(active) != 0 {
		t.Fatalf("expected 0 active sessions after end, got %d", len(active))
	}
}

func TestFileTouchedRequiresKnownSession(t *testing.T) {
	r := New()

	err := r.FileTouched("sess-unknown", "src/foo.ts", "write", time.Now())
	if err == nil {
		t.Fatal("expected error for unknown session, got nil")
	}
}

func TestFileTouchedKnownSession(t *testing.T) {
	r := New()
	now := time.Now()

	r.SessionStarted("sess-a", "/repo", "Alice", "claude-code", now)

	if err := r.FileTouched("sess-a", "src/foo.ts", "write", now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	signals := r.Query("/repo", []string{"src/foo.ts"}, now, time.Hour)
	if len(signals["src/foo.ts"]) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals["src/foo.ts"]))
	}
}

func TestQueryRespectsWindow(t *testing.T) {
	r := New()
	start := time.Now()

	r.SessionStarted("sess-a", "/repo", "Alice", "claude-code", start)
	if err := r.FileTouched("sess-a", "src/foo.ts", "write", start); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := start.Add(2 * time.Hour)

	tests := []struct {
		name   string
		window time.Duration
		want   int
	}{
		{"within window", 3 * time.Hour, 1},
		{"outside window", time.Hour, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signals := r.Query("/repo", []string{"src/foo.ts"}, now, tt.window)
			if got := len(signals["src/foo.ts"]); got != tt.want {
				t.Fatalf("got %d signals, want %d", got, tt.want)
			}
		})
	}
}

func TestQueryReturnsEntryForEveryRequestedPath(t *testing.T) {
	r := New()
	now := time.Now()

	r.SessionStarted("sess-a", "/repo", "Alice", "claude-code", now)
	if err := r.FileTouched("sess-a", "src/foo.ts", "write", now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	signals := r.Query("/repo", []string{"src/foo.ts", "src/bar.ts"}, now, time.Hour)

	if _, ok := signals["src/foo.ts"]; !ok {
		t.Fatal("expected entry for src/foo.ts")
	}

	barSignals, ok := signals["src/bar.ts"]
	if !ok {
		t.Fatal("expected entry for src/bar.ts")
	}
	if len(barSignals) != 0 {
		t.Fatalf("expected empty signals for untouched path, got %d", len(barSignals))
	}
}

func TestActiveSessionsExcludesEnded(t *testing.T) {
	r := New()
	now := time.Now()

	r.SessionStarted("sess-a", "/repo", "Alice", "claude-code", now)
	r.SessionStarted("sess-b", "/repo", "Bob", "opencode", now)
	r.SessionEnded("sess-b", now)

	active := r.ActiveSessions("/repo")
	if len(active) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(active))
	}
	if active[0].SessionID != "sess-a" {
		t.Fatalf("expected sess-a active, got %s", active[0].SessionID)
	}
}

func TestSweepRemovesStaleEntries(t *testing.T) {
	r := New()
	start := time.Now()

	r.SessionStarted("sess-a", "/repo", "Alice", "claude-code", start)
	if err := r.FileTouched("sess-a", "src/foo.ts", "write", start); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r.SessionEnded("sess-a", start)

	future := start.Add(time.Hour)
	r.Sweep(future, 30*time.Minute)

	if len(r.sessions) != 0 {
		t.Fatalf("expected sessions to be swept, got %d", len(r.sessions))
	}

	signals := r.Query("/repo", []string{"src/foo.ts"}, future, time.Hour)
	if len(signals["src/foo.ts"]) != 0 {
		t.Fatalf("expected touches to be swept, got %d", len(signals["src/foo.ts"]))
	}
}

func TestActiveSessionsIncludesHarness(t *testing.T) {
	r := New()
	now := time.Now()

	r.SessionStarted("sess-a", "/repo", "Alice", "opencode", now)

	active := r.ActiveSessions("/repo")
	if len(active) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(active))
	}
	if active[0].Harness != "opencode" {
		t.Fatalf("expected harness %q, got %q", "opencode", active[0].Harness)
	}
}

func TestSessionStartedDefaultsHarnessWhenEmpty(t *testing.T) {
	r := New()
	now := time.Now()

	r.SessionStarted("sess-a", "/repo", "Alice", "", now)

	active := r.ActiveSessions("/repo")
	if len(active) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(active))
	}
	if active[0].Harness != "unknown" {
		t.Fatalf("expected harness %q, got %q", "unknown", active[0].Harness)
	}
}

func TestSweepKeepsActiveEntries(t *testing.T) {
	r := New()
	start := time.Now()

	r.SessionStarted("sess-a", "/repo", "Alice", "claude-code", start)
	if err := r.FileTouched("sess-a", "src/foo.ts", "write", start); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	soon := start.Add(time.Minute)
	r.Sweep(soon, 30*time.Minute)

	if len(r.sessions) != 1 {
		t.Fatalf("expected session to survive sweep, got %d", len(r.sessions))
	}
}
