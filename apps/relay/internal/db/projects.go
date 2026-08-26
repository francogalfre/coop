package db

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/project"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/projectinvite"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/projectmember"
)

const (
	RoleOwner  = "owner"
	RoleMember = "member"
)

func (p *Pool) CreateProject(ctx context.Context, name, slug, ownerID string) (*ent.Project, error) {
	tx, err := p.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: create project: begin: %w", err)
	}

	proj, err := tx.Project.Create().
		SetName(name).
		SetSlug(slug).
		SetCreatedBy(ownerID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("db: create project: %w", err)
	}

	if _, err := tx.ProjectMember.Create().
		SetProject(proj).
		SetUserID(ownerID).
		SetRole(RoleOwner).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("db: create project: add owner: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("db: create project: commit: %w", err)
	}

	return proj, nil
}

func (p *Pool) GetProjectBySlug(ctx context.Context, slug string) (*ent.Project, error) {
	proj, err := p.client.Project.Query().Where(project.Slug(slug)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: get project: %w", err)
	}

	return proj, nil
}

func (p *Pool) MemberRole(ctx context.Context, projectID int, userID string) (role string, isMember bool, err error) {
	m, err := p.client.ProjectMember.Query().
		Where(projectmember.UserID(userID), projectmember.HasProjectWith(project.ID(projectID))).
		Only(ctx)
	if IsNotFound(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("db: member role: %w", err)
	}

	return m.Role, true, nil
}

func (p *Pool) AddMember(ctx context.Context, proj *ent.Project, userID, role string) error {
	exists, err := p.client.ProjectMember.Query().
		Where(projectmember.UserID(userID), projectmember.HasProjectWith(project.ID(proj.ID))).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("db: add member: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := p.client.ProjectMember.Create().
		SetProject(proj).
		SetUserID(userID).
		SetRole(role).
		Save(ctx); err != nil {
		return fmt.Errorf("db: add member: %w", err)
	}

	return nil
}

const inviteTokenBytes = 24

func (p *Pool) CreateInvite(ctx context.Context, proj *ent.Project, createdBy string, ttl time.Duration) (rawToken string, err error) {
	raw := make([]byte, inviteTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("db: create invite: generate token: %w", err)
	}

	rawToken = hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(rawToken))

	if _, err := p.client.ProjectInvite.Create().
		SetProject(proj).
		SetTokenHash(sum[:]).
		SetCreatedBy(createdBy).
		SetExpiresAt(time.Now().Add(ttl)).
		Save(ctx); err != nil {
		return "", fmt.Errorf("db: create invite: %w", err)
	}

	return rawToken, nil
}

func (p *Pool) AcceptInvite(ctx context.Context, rawToken, userID string) (*ent.Project, error) {
	sum := sha256.Sum256([]byte(rawToken))

	inv, err := p.client.ProjectInvite.Query().
		Where(
			projectinvite.TokenHash(sum[:]),
			projectinvite.RevokedAtIsNil(),
			projectinvite.ExpiresAtGT(time.Now()),
		).
		WithProject().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: accept invite: %w", err)
	}

	proj := inv.Edges.Project

	if err := p.AddMember(ctx, proj, userID, RoleMember); err != nil {
		return nil, err
	}

	return proj, nil
}
