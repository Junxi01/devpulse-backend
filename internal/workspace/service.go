package workspace

import (
	"context"
	"errors"
	"strings"

	"devpulse-backend/internal/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidName   = errors.New("invalid workspace name")
	ErrNotFound      = errors.New("workspace not found")
	ErrUnauthenticated = errors.New("unauthenticated")
)

type Service struct {
	Pool *pgxpool.Pool
	Q    *generated.Queries
}

func (s Service) Create(ctx context.Context, userID uuid.UUID, name string) (generated.Workspace, error) {
	if userID == uuid.Nil {
		return generated.Workspace{}, ErrUnauthenticated
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return generated.Workspace{}, ErrInvalidName
	}
	if s.Pool == nil || s.Q == nil {
		return generated.Workspace{}, errors.New("service unavailable")
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return generated.Workspace{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := s.Q.WithTx(tx)
	ws, err := qtx.CreateWorkspace(ctx, generated.CreateWorkspaceParams{
		Name:    name,
		OwnerID: userID,
	})
	if err != nil {
		return generated.Workspace{}, err
	}
	if err := qtx.AddWorkspaceMember(ctx, generated.AddWorkspaceMemberParams{
		WorkspaceID: ws.ID,
		UserID:      userID,
		Role:        "owner",
	}); err != nil {
		return generated.Workspace{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return generated.Workspace{}, err
	}
	return ws, nil
}

func (s Service) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]generated.Workspace, error) {
	if userID == uuid.Nil {
		return nil, ErrUnauthenticated
	}
	if s.Q == nil {
		return nil, errors.New("service unavailable")
	}
	if limit <= 0 {
		limit = 50
	}
	return s.Q.ListWorkspacesByUser(ctx, generated.ListWorkspacesByUserParams{
		UserID:  userID,
		Limit:   limit,
		Offset:  offset,
	})
}

func (s Service) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (generated.Workspace, error) {
	if userID == uuid.Nil {
		return generated.Workspace{}, ErrUnauthenticated
	}
	if s.Q == nil {
		return generated.Workspace{}, errors.New("service unavailable")
	}
	ws, err := s.Q.GetWorkspaceByID(ctx, generated.GetWorkspaceByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return generated.Workspace{}, ErrNotFound
		}
		return generated.Workspace{}, err
	}
	return ws, nil
}

