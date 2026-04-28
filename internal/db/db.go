package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	SQL *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*DB, error) {
	d, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	// Conservative defaults; can be tuned later.
	d.SetMaxOpenConns(10)
	d.SetMaxIdleConns(10)
	d.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := d.PingContext(pingCtx); err != nil {
		_ = d.Close()
		return nil, err
	}

	return &DB{SQL: d}, nil
}

