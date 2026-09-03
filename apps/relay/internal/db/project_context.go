package db

import (
	"context"
	"fmt"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
)

func (p *Pool) SetProjectContext(ctx context.Context, projectID int, text, updatedBy string) (*ent.Project, error) {
	proj, err := p.client.Project.UpdateOneID(projectID).
		SetContextText(text).
		SetContextUpdatedBy(updatedBy).
		SetContextUpdatedAt(time.Now()).
		AddContextVersion(1).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: set project context: %w", err)
	}

	return proj, nil
}
