package db

import (
	"context"
	"fmt"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/agentsession"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/steerrequest"
)

func (p *Pool) CreateSteerRequest(ctx context.Context, requestID, sessionID, actorID, actorDisplayName, actorAvatarURL, text string) error {
	err := p.client.SteerRequest.Create().
		SetID(requestID).
		SetSessionID(sessionID).
		SetActorID(actorID).
		SetActorDisplayName(actorDisplayName).
		SetActorAvatarURL(actorAvatarURL).
		SetText(text).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("db: create steer request: %w", err)
	}

	return nil
}

func (p *Pool) GetSteerRequest(ctx context.Context, sessionID, requestID string) (*ent.SteerRequest, error) {
	req, err := p.client.SteerRequest.Query().
		Where(steerrequest.ID(requestID), steerrequest.HasSessionWith(agentsession.ID(sessionID))).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: get steer request: %w", err)
	}

	return req, nil
}

func (p *Pool) DeleteSteerRequest(ctx context.Context, requestID string) error {
	err := p.client.SteerRequest.DeleteOneID(requestID).Exec(ctx)
	if err != nil && !IsNotFound(err) {
		return fmt.Errorf("db: delete steer request: %w", err)
	}

	return nil
}

// Mirrors the in-memory registry's oldest-dropped eviction so the cap holds across a restart too.
func (p *Pool) EvictOldestSteerRequests(ctx context.Context, sessionID string, keep int) ([]string, error) {
	ids, err := p.client.SteerRequest.Query().
		Where(steerrequest.HasSessionWith(agentsession.ID(sessionID))).
		Order(ent.Asc(steerrequest.FieldCreatedAt)).
		IDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: evict steer requests: list: %w", err)
	}

	if len(ids) <= keep {
		return nil, nil
	}

	dropped := ids[:len(ids)-keep]
	if _, err := p.client.SteerRequest.Delete().Where(steerrequest.IDIn(dropped...)).Exec(ctx); err != nil {
		return nil, fmt.Errorf("db: evict steer requests: delete: %w", err)
	}

	return dropped, nil
}
