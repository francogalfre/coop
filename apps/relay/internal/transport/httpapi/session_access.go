package httpapi

import (
	"context"

	"github.com/francogalfre/coop/apps/relay/internal/db"
)

func memberSessionIDs(ctx context.Context, pool *db.Pool, userID string) (map[string]bool, error) {
	projects, err := pool.ListUserProjects(ctx, userID)
	if err != nil {
		return nil, err
	}

	ids := map[string]bool{}
	for _, proj := range projects {
		sessions, err := pool.ListProjectSessions(ctx, proj.ID)
		if err != nil {
			return nil, err
		}

		for _, sess := range sessions {
			ids[sess.ID] = true
		}
	}

	return ids, nil
}
