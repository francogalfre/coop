package hooks

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

// translateEvent picks the per-harness payload translator. claude-code (and
// anything unrecognized reaching this endpoint) goes through buildEventBody,
// the only translator with fully verified field shapes end to end.
func translateEvent(harnessName string, nextSeq func() int, sessionID, event string, payload map[string]any, red *redact.Redactor) ([][]byte, error) {
	switch harnessName {
	case "opencode":
		return translateOpencode(nextSeq, sessionID, event, payload, red)
	case "pi":
		return translatePi(nextSeq, sessionID, event, payload, red)
	default:
		return buildEventBody(nextSeq, sessionID, event, payload, red)
	}
}

func emitEnvelope(bodies *[][]byte, nextSeq func() int, sessionID string, fields map[string]any) error {
	envelope := map[string]any{
		"v":          1,
		"session_id": sessionID,
		"seq":        nextSeq(),
		"ts":         time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range fields {
		envelope[k] = v
	}

	out, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("hooks: marshal event body: %w", err)
	}

	*bodies = append(*bodies, out)

	return nil
}
