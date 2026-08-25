package relayclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

const requestTimeout = 10 * time.Second

var httpClient = &http.Client{Timeout: requestTimeout}

func PostEvent(ctx context.Context, cfg config.Config, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.RelayURL+"/v1/events", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("relayclient: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("relayclient: post event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("relayclient: post event: unexpected status %d: %s", resp.StatusCode, readBody(resp.Body))
	}

	return nil
}

type steerMessageBody struct {
	From string `json:"from"`
	Text string `json:"text"`
}

func GetSteer(ctx context.Context, cfg config.Config, sessionID string) (from, text string, ok bool, err error) {
	steerURL := cfg.RelayURL + "/v1/sessions/" + url.PathEscape(sessionID) + "/steer"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, steerURL, nil)
	if err != nil {
		return "", "", false, fmt.Errorf("relayclient: build request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", false, fmt.Errorf("relayclient: get steer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return "", "", false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", false, fmt.Errorf("relayclient: get steer: unexpected status %d: %s", resp.StatusCode, readBody(resp.Body))
	}

	var msg steerMessageBody
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return "", "", false, fmt.Errorf("relayclient: decode steer response: %w", err)
	}

	return msg.From, msg.Text, true, nil
}

func readBody(r io.Reader) string {
	b, err := io.ReadAll(io.LimitReader(r, 4096))
	if err != nil {
		return ""
	}

	return string(b)
}
