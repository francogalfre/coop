package db

import (
	"context"
	"fmt"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/agentsession"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/event"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/project"
)

const (
	SessionStatusLive  = "live"
	SessionStatusEnded = "ended"

	SessionModeAuto       = "auto"
	SessionModeRestricted = "restricted"
)

func (p *Pool) CreateAgentSession(ctx context.Context, id string, proj *ent.Project, ownerID, repo, cwd, harness string, startedAt time.Time) (*ent.AgentSession, error) {
	sess, err := p.client.AgentSession.Create().
		SetID(id).
		SetProject(proj).
		SetOwnerID(ownerID).
		SetRepo(repo).
		SetCwd(cwd).
		SetHarness(harness).
		SetStatus(SessionStatusLive).
		SetStartedAt(startedAt).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: create agent session: %w", err)
	}

	return sess, nil
}

func (p *Pool) EndAgentSession(ctx context.Context, id string, endedAt time.Time) error {
	_, err := p.client.AgentSession.UpdateOneID(id).
		SetStatus(SessionStatusEnded).
		SetEndedAt(endedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("db: end agent session: %w", err)
	}

	return nil
}

func (p *Pool) EndStaleSessions(ctx context.Context, idleSince time.Time) ([]string, error) {
	ids, err := p.client.AgentSession.Query().
		Where(
			agentsession.StatusEQ(SessionStatusLive),
			agentsession.StartedAtLT(idleSince),
			agentsession.Not(agentsession.HasEventsWith(event.CreatedAtGTE(idleSince))),
		).
		IDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: end stale sessions: find: %w", err)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	if err := p.client.AgentSession.Update().
		Where(agentsession.IDIn(ids...)).
		SetStatus(SessionStatusEnded).
		SetEndedAt(time.Now().UTC()).
		Exec(ctx); err != nil {
		return nil, fmt.Errorf("db: end stale sessions: update: %w", err)
	}

	return ids, nil
}

func (p *Pool) GetAgentSession(ctx context.Context, id string) (*ent.AgentSession, error) {
	sess, err := p.client.AgentSession.Query().
		Where(agentsession.ID(id)).
		WithProject().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: get agent session: %w", err)
	}

	return sess, nil
}

func (p *Pool) SetSessionMode(ctx context.Context, id string, mode string) (*ent.AgentSession, error) {
	sess, err := p.client.AgentSession.UpdateOneID(id).
		SetMode(mode).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: set session mode: %w", err)
	}

	return sess, nil
}

func (p *Pool) ListProjectSessions(ctx context.Context, projectID int) ([]*ent.AgentSession, error) {
	sessions, err := p.client.AgentSession.Query().
		Where(agentsession.HasProjectWith(project.ID(projectID))).
		Order(ent.Desc(agentsession.FieldStartedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: list project sessions: %w", err)
	}

	return sessions, nil
}
