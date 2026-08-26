package db

import "github.com/francogalfre/coop/apps/relay/internal/db/ent"

func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}
