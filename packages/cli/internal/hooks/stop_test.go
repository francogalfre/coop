package hooks

import (
	"strings"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

func TestBuildEventBodyStopEmitsTurnEndAndAgentText(t *testing.T) {
	payload := map[string]any{
		"hook_event_name":        "Stop",
		"last_assistant_message": "done, key sk-abcdefghijklmnopqrstuv",
	}

	bodies, err := buildEventBody(config.Config{}, seqFrom(4), "sess-a", "Stop", payload, redact.New())
	if err != nil {
		t.Fatalf("buildEventBody() error = %v", err)
	}

	types := eventTypes(t, bodies)
	if len(types) != 2 || types[0] != "agent.turn_end" || types[1] != "agent.text" {
		t.Fatalf("got types %v, want [agent.turn_end agent.text]", types)
	}

	text := decodeBody(t, bodies[1])
	textField, ok := text["text"].(map[string]any)
	if !ok {
		t.Fatalf("got text %v, want an object", text["text"])
	}
	if strings.Contains(textField["text"].(string), "sk-abcdefghijklmnopqrstuv") {
		t.Errorf("agent.text still contains the secret: %v", textField["text"])
	}
}

func TestBuildEventBodySessionEndReason(t *testing.T) {
	tests := []struct {
		reason     string
		wantReason bool
	}{
		{"completed", true},
		{"cancelled", true},
		{"error", true},
		{"other", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			payload := map[string]any{"hook_event_name": "SessionEnd", "reason": tt.reason}

			out := buildOne(t, 3, "sess-a", "SessionEnd", payload, redact.New())
			if out["type"] != "session.end" {
				t.Errorf("got type %v, want session.end", out["type"])
			}

			_, hasReason := out["reason"]
			if hasReason != tt.wantReason {
				t.Errorf("reason=%q: got hasReason=%v, want %v", tt.reason, hasReason, tt.wantReason)
			}
			if tt.wantReason && out["reason"] != tt.reason {
				t.Errorf("got reason %v, want %v", out["reason"], tt.reason)
			}
		})
	}
}
