package db

import (
	"context"
	"database/sql"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
)

type Pool struct {
	client *ent.Client
	sqlDB  *sql.DB
}

func Open(ctx context.Context, connString string) (*Pool, error) {
	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	driver := entsql.OpenDB("postgres", sqlDB)
	client := ent.NewClient(ent.Driver(driver))

	return &Pool{client: client, sqlDB: sqlDB}, nil
}

func (p *Pool) Close() error {
	return p.client.Close()
}

func (p *Pool) Ping(ctx context.Context) error {
	return p.sqlDB.PingContext(ctx)
}

func (p *Pool) Migrate(ctx context.Context) error {
	return p.client.Schema.Create(ctx)
}
