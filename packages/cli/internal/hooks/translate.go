package hooks

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/redact"
)

// claude-code (and anything unrecognized) goes through buildEventBody, the only translator with fully verified field shapes end to end.
func translateEvent(harnessName string, cfg config.Config, nextSeq func() int, sessionID, event string, payload map[string]any, red *redact.Redactor) ([][]byte, error) {
	switch harnessName {
	case "opencode":
		return translateOpencode(cfg, nextSeq, sessionID, event, payload, red)
	case "pi":
		return translatePi(cfg, nextSeq, sessionID, event, payload, red)
	default:
		return buildEventBody(cfg, nextSeq, sessionID, event, payload, red)
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
