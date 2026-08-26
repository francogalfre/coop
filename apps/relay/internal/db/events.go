package db

import (
	"context"
	"fmt"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/agentsession"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/event"
)

func (p *Pool) AppendEvent(ctx context.Context, sessionID string, data []byte) (*ent.Event, error) {
	tx, err := p.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: append event: begin: %w", err)
	}

	sess, err := tx.AgentSession.UpdateOneID(sessionID).AddNextSeq(1).Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("db: append event: reserve seq: %w", err)
	}

	e, err := tx.Event.Create().
		SetSession(sess).
		SetSeq(sess.NextSeq).
		SetData(data).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("db: append event: insert: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("db: append event: commit: %w", err)
	}

	return e, nil
}

func (p *Pool) EventsSince(ctx context.Context, sessionID string, afterSeq, limit int) ([]*ent.Event, error) {
	events, err := p.client.Event.Query().
		Where(event.HasSessionWith(agentsession.ID(sessionID)), event.SeqGT(afterSeq)).
		Order(ent.Asc(event.FieldSeq)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: events since: %w", err)
	}

	return events, nil
}
