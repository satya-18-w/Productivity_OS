// Package postgres owns the PostgreSQL connection pool and its lifecycle. It holds
// no business logic (ADR-0003).
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open builds a pgx pool from a connection string and verifies connectivity. A
// failure here is fatal for the process (ADR-0003: verify before serving).
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database unreachable: %w", err)
	}
	return pool, nil
}

// Healthy reports whether the pool can currently reach the database. Used by the
// readiness check.
func Healthy(ctx context.Context, pool *pgxpool.Pool) bool {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return pool.Ping(pingCtx) == nil
}
