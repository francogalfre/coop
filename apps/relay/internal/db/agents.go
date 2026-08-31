package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/agent"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/agentsession"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/project"
)

const agentIDBytes = 8

func (p *Pool) CreateAgent(ctx context.Context, projectID int, name, displayName, createdBy string) (*ent.Agent, error) {
	existing, err := p.GetAgentByName(ctx, projectID, name)
	if err == nil {
		return existing, nil
	}
	if !IsNotFound(err) {
		return nil, err
	}

	id, err := generateAgentID()
	if err != nil {
		return nil, fmt.Errorf("db: create agent: %w", err)
	}

	created, err := p.client.Agent.Create().
		SetID(id).
		SetProjectID(projectID).
		SetName(name).
		SetDisplayName(displayName).
		SetCreatedBy(createdBy).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return p.GetAgentByName(ctx, projectID, name)
		}

		return nil, fmt.Errorf("db: create agent: %w", err)
	}

	return created, nil
}

func (p *Pool) GetAgentByName(ctx context.Context, projectID int, name string) (*ent.Agent, error) {
	a, err := p.client.Agent.Query().
		Where(agent.Name(name), agent.HasProjectWith(project.ID(projectID))).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: get agent: %w", err)
	}

	return a, nil
}

func (p *Pool) ListProjectAgents(ctx context.Context, projectID int) ([]*ent.Agent, error) {
	agents, err := p.client.Agent.Query().
		Where(agent.HasProjectWith(project.ID(projectID))).
		Order(ent.Asc(agent.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: list project agents: %w", err)
	}

	return agents, nil
}

func (p *Pool) CurrentSessionForAgent(ctx context.Context, agentID string) (*ent.AgentSession, error) {
	sess, err := p.client.AgentSession.Query().
		Where(agentsession.HasAgentWith(agent.ID(agentID)), agentsession.StatusEQ(SessionStatusLive)).
		Order(ent.Desc(agentsession.FieldStartedAt)).
		First(ctx)
	if err != nil {
		if IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("db: current session for agent: %w", err)
	}

	return sess, nil
}

func (p *Pool) LinkSessionToAgent(ctx context.Context, sessionID, agentID string) error {
	if err := p.client.AgentSession.UpdateOneID(sessionID).SetAgentID(agentID).Exec(ctx); err != nil {
		return fmt.Errorf("db: link session to agent: %w", err)
	}

	return nil
}

func generateAgentID() (string, error) {
	raw := make([]byte, agentIDBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate agent id: %w", err)
	}

	return "agent-" + hex.EncodeToString(raw), nil
}
