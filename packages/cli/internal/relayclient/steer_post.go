package relayclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

type steerPostBody struct {
	Text                  string `json:"text"`
	ProjectContextVersion *int   `json:"project_context_version,omitempty"`
}

func DeliverSteer(ctx context.Context, cfg config.Config, sessionID, text string, contextVersion *int) error {
	payload, err := json.Marshal(steerPostBody{Text: text, ProjectContextVersion: contextVersion})
	if err != nil {
		return fmt.Errorf("relayclient: deliver steer: encode request: %w", err)
	}

	steerURL := cfg.RelayURL + "/v1/sessions/" + url.PathEscape(sessionID) + "/steer"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, steerURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("relayclient: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	setCLICredential(req, cfg)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("relayclient: deliver steer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("relayclient: deliver steer: unexpected status %d: %s", resp.StatusCode, readBody(resp.Body))
	}

	return nil
}
