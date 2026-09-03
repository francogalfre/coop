package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/project"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent/projectnote"
)

const projectNoteIDBytes = 12

func (p *Pool) CreateProjectNote(ctx context.Context, projectID int, authorID, authorDisplayName, authorAvatarURL, source, sessionID, text string) (*ent.ProjectNote, error) {
	id, err := generateProjectNoteID()
	if err != nil {
		return nil, fmt.Errorf("db: create project note: %w", err)
	}

	note, err := p.client.ProjectNote.Create().
		SetID(id).
		SetProjectID(projectID).
		SetAuthorID(authorID).
		SetAuthorDisplayName(authorDisplayName).
		SetAuthorAvatarURL(authorAvatarURL).
		SetSource(source).
		SetSessionID(sessionID).
		SetText(text).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: create project note: %w", err)
	}

	return note, nil
}

const projectNotesDefaultLimit = 50

func (p *Pool) ListProjectNotes(ctx context.Context, projectID int, limit int) ([]*ent.ProjectNote, error) {
	if limit <= 0 || limit > projectNotesDefaultLimit {
		limit = projectNotesDefaultLimit
	}

	notes, err := p.client.ProjectNote.Query().
		Where(projectnote.HasProjectWith(project.ID(projectID))).
		Order(ent.Desc(projectnote.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: list project notes: %w", err)
	}

	return notes, nil
}

func generateProjectNoteID() (string, error) {
	raw := make([]byte, projectNoteIDBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate project note id: %w", err)
	}

	return "note-" + hex.EncodeToString(raw), nil
}
