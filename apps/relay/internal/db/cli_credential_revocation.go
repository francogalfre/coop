package db

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent/clicredential"
)

// Deleting zero rows (a token already revoked, or one that never matched userID) is treated as success so revocation stays idempotent and never leaks which case occurred.
func (p *Pool) RevokeCliCredential(ctx context.Context, userID string, rawToken []byte) error {
	sum := sha256.Sum256(rawToken)

	if _, err := p.client.CliCredential.Delete().
		Where(
			clicredential.UserID(userID),
			clicredential.TokenHash(sum[:]),
		).
		Exec(ctx); err != nil {
		return fmt.Errorf("db: revoke cli credential: %w", err)
	}

	return nil
}
