package repository

import (
	"context"

	"devpulse-backend/internal/db/generated"

	"github.com/google/uuid"
)

// UserRepository is a thin wrapper around sqlc queries.
// It exists so higher layers (e.g. auth) depend on a small interface rather than sqlc directly.
type UserRepository interface {
	Create(ctx context.Context, email string, passwordHash string, name string) (generated.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (generated.User, error)
	GetByEmail(ctx context.Context, email string) (generated.User, error)
}

type userRepo struct {
	q *generated.Queries
}

func NewUserRepository(q *generated.Queries) UserRepository {
	return &userRepo{q: q}
}

func (r *userRepo) Create(ctx context.Context, email string, passwordHash string, name string) (generated.User, error) {
	return r.q.CreateUser(ctx, generated.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
	})
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (generated.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (generated.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

