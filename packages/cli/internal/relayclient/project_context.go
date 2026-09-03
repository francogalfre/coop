package relayclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

type ProjectContext struct {
	Text    string
	Version int
}

type projectContextGetBody struct {
	Text    string `json:"text"`
	Version int    `json:"version"`
}

func GetProjectContext(ctx context.Context, cfg config.Config, project string) (ProjectContext, error) {
	contextURL := cfg.RelayURL + "/v1/projects/" + url.PathEscape(project) + "/context"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, contextURL, nil)
	if err != nil {
		return ProjectContext{}, fmt.Errorf("relayclient: build request: %w", err)
	}

	setCLICredential(req, cfg)

	resp, err := httpClient.Do(req)
	if err != nil {
		return ProjectContext{}, fmt.Errorf("relayclient: get project context: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ProjectContext{}, fmt.Errorf("relayclient: get project context: unexpected status %d: %s", resp.StatusCode, readBody(resp.Body))
	}

	var body projectContextGetBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return ProjectContext{}, fmt.Errorf("relayclient: decode project context response: %w", err)
	}

	return ProjectContext{Text: body.Text, Version: body.Version}, nil
}
