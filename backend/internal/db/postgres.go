package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier describes the query methods shared by pgxpool.Pool and pgx.Tx so
// repositories can stay storage-agnostic and easy to test.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

// Database wraps the shared pgx connection pool used by repositories.
type Database struct {
	pool *pgxpool.Pool
}

// New opens a pgx connection pool and verifies database reachability with a
// startup ping.
func New(ctx context.Context, databaseURL string) (*Database, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse pgx pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Database{pool: pool}, nil
}

// Close shuts down the underlying connection pool.
func (d *Database) Close() {
	if d == nil || d.pool == nil {
		return
	}

	d.pool.Close()
}

// Pool exposes the underlying pool for cases that need pgx-specific features.
func (d *Database) Pool() *pgxpool.Pool {
	if d == nil {
		return nil
	}

	return d.pool
}

// Querier returns the pool as a repository-compatible query interface.
func (d *Database) Querier() Querier {
	if d == nil {
		return nil
	}

	return d.pool
}

// Begin starts a transaction from the shared pool.
func (d *Database) Begin(ctx context.Context) (pgx.Tx, error) {
	if d == nil || d.pool == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	return d.pool.Begin(ctx)
}
