package hooks

import (
	"github.com/francogalfre/coop/packages/cli/internal/redact"
	"github.com/francogalfre/coop/packages/cli/internal/repoid"
)

func translateOpencode(nextSeq func() int, sessionID, event string, payload map[string]any, red *redact.Redactor) ([][]byte, error) {
	var bodies [][]byte

	switch event {
	case "session.created":
		rawDir := stringField(payload, "directory")
		dir, _ := red.Text(rawDir)
		if dir == "" {
			return bodies, emitGenericOpencode(&bodies, nextSeq, sessionID, event, payload, red)
		}

		fields := map[string]any{
			"type":    "session.start",
			"harness": "opencode",
			"cwd":     dir,
			"owner":   actorFields(),
		}

		if repo := repoid.Detect(rawDir); repo != "" {
			fields["repo"] = repo
		}

		return bodies, emitEnvelope(&bodies, nextSeq, sessionID, fields)

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
