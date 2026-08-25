package hooks

import "github.com/francogalfre/coop/packages/cli/internal/redact"

// translateOpencode covers the two fields the opencode plugin API actually
// documents to us: the plugin factory's own "directory" argument, and the
// bus event's "type" discriminator. Everything else (tool args/output,
// file.edited's path) has no verified field shape yet, so it round-trips
// as a redacted generic event instead of guessing field names.
func translateOpencode(nextSeq func() int, sessionID, event string, payload map[string]any, red *redact.Redactor) ([][]byte, error) {
	var bodies [][]byte

	switch event {
	case "session.created":
		dir, _ := red.Text(stringField(payload, "directory"))
		if dir == "" {
			return bodies, emitGenericOpencode(&bodies, nextSeq, sessionID, event, payload, red)
		}

		return bodies, emitEnvelope(&bodies, nextSeq, sessionID, map[string]any{
			"type":    "session.start",
			"harness": "opencode",
			"cwd":     dir,
			"owner":   actorFields(),
		})

	case "session.idle":
		return bodies, emitEnvelope(&bodies, nextSeq, sessionID, map[string]any{"type": "agent.turn_end"})

	default:
		return bodies, emitGenericOpencode(&bodies, nextSeq, sessionID, event, payload, red)
	}
}

func emitGenericOpencode(bodies *[][]byte, nextSeq func() int, sessionID, event string, payload map[string]any, red *redact.Redactor) error {
	redactedPayload, _ := red.Value(payload)

	return emitEnvelope(bodies, nextSeq, sessionID, map[string]any{
		"type": "hook.opencode." + event,
		"raw":  redactedPayload,
	})
}
