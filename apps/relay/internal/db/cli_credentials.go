package db

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent/clicredential"
)

const credentialTokenBytes = 32

func (p *Pool) CreateCliCredential(ctx context.Context, userID, displayName string) (rawToken []byte, err error) {
	raw := make([]byte, credentialTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("db: create cli credential: generate token: %w", err)
	}

	sum := sha256.Sum256(raw)

	if _, err := p.client.CliCredential.Create().
		SetUserID(userID).
		SetDisplayName(displayName).
		SetTokenHash(sum[:]).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("db: create cli credential: %w", err)
	}

	return raw, nil
}

func (p *Pool) AuthenticateCliCredential(ctx context.Context, rawToken []byte) (userID, displayName string, err error) {
	sum := sha256.Sum256(rawToken)

	cred, err := p.client.CliCredential.Query().
		Where(
			clicredential.TokenHash(sum[:]),
			clicredential.Or(
				clicredential.ExpiresAtIsNil(),
				clicredential.ExpiresAtGT(time.Now()),
			),
		).
		Only(ctx)
	if err != nil {
		return "", "", fmt.Errorf("db: authenticate cli credential: %w", err)
	}

	if _, err := cred.Update().SetLastUsedAt(time.Now()).Save(ctx); err != nil {
		return "", "", fmt.Errorf("db: authenticate cli credential: touch: %w", err)
	}

	return cred.UserID, cred.DisplayName, nil
}
