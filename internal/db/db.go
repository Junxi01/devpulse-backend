package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"devpulse-backend/internal/db/generated"
)

type DB struct {
	Pool    *pgxpool.Pool
	Queries *generated.Queries
}

func Open(ctx context.Context, databaseURL string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Conservative defaults; can be tuned later.
	cfg.MaxConns = 10
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 30 * time.Minute

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return &DB{
		Pool:    pool,
		Queries: generated.New(pool),
	}, nil
}

