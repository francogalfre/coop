package db

import (
	"context"
	"fmt"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
)

func (p *Pool) GetTakeover(ctx context.Context, sessionID string) (*ent.Takeover, error) {
	row, err := p.client.Takeover.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("db: get takeover: %w", err)
	}

	return row, nil
}

// No upsert here: ent's sql/upsert feature isn't enabled, so a claim replaces any prior row via delete-then-create.
func (p *Pool) SetTakeover(ctx context.Context, sessionID, actorID, actorDisplayName string) error {
	if err := p.client.Takeover.DeleteOneID(sessionID).Exec(ctx); err != nil && !IsNotFound(err) {
		return fmt.Errorf("db: set takeover: clear existing: %w", err)
	}

	err := p.client.Takeover.Create().
		SetID(sessionID).
		SetSessionID(sessionID).
		SetActorID(actorID).
		SetActorDisplayName(actorDisplayName).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("db: set takeover: %w", err)
	}

	return nil
}

func (p *Pool) ClearTakeover(ctx context.Context, sessionID string) error {
	if err := p.client.Takeover.DeleteOneID(sessionID).Exec(ctx); err != nil && !IsNotFound(err) {
		return fmt.Errorf("db: clear takeover: %w", err)
	}

	return nil
}
