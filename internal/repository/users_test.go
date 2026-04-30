package repository

import (
	"context"
	"testing"

	"devpulse-backend/internal/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type stubDBTX struct{}

func (stubDBTX) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (stubDBTX) Query(context.Context, string, ...interface{}) (pgx.Rows, error) { return nil, nil }
func (stubDBTX) QueryRow(context.Context, string, ...interface{}) pgx.Row          { return nil }

func TestUserRepository_Compiles(t *testing.T) {
	q := generated.New(stubDBTX{})
	repo := NewUserRepository(q)

	// Compile-time assertions: do not execute DB calls in this test.
	var _ UserRepository = repo
	_ = uuid.Nil
	_ = context.Background()
}

