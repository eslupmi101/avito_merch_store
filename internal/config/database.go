package config

import (
	"context"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDb struct {
	Connection *pgxpool.Pool
}

var (
	pgInstance *PostgresDb
	pgOnce     sync.Once
)

func NewPostgresDb(ctx context.Context, connStr string) *PostgresDb {
	pgOnce.Do(
		func() {
			db, err := pgxpool.New(ctx, connStr)
			if err != nil {
				log.Fatalf("Unable to create connection pool: %s", err)
			}
			pgInstance = &PostgresDb{db}
		},
	)

	return pgInstance
}

func (db *PostgresDb) Ping(ctx context.Context) error {
	return db.Connection.Ping(ctx)
}

func (db *PostgresDb) Close() {
	db.Connection.Close()
}
