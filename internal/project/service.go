package project

import (
	"context"
	"errors"
	"strings"

	"devpulse-backend/internal/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidName     = errors.New("invalid project name")
	ErrConflict        = errors.New("conflict")
)

type Service struct {
	Q *generated.Queries
}

func (s Service) ensureWorkspaceMember(ctx context.Context, userID, workspaceID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrUnauthenticated
	}
	if s.Q == nil {
		return errors.New("service unavailable")
	}
	ok, err := s.Q.IsUserWorkspaceMember(ctx, generated.IsUserWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s Service) CreateForWorkspace(ctx context.Context, userID, workspaceID uuid.UUID, name, description string) (generated.Project, error) {
	if err := s.ensureWorkspaceMember(ctx, userID, workspaceID); err != nil {
		return generated.Project{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return generated.Project{}, ErrInvalidName
	}
	desc := strings.TrimSpace(description)
	p, err := s.Q.CreateProject(ctx, generated.CreateProjectParams{
		WorkspaceID: workspaceID,
		Name:        name,
		Description: desc,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return generated.Project{}, ErrConflict
		}
		return generated.Project{}, err
	}
	return p, nil
}

func (s Service) ListForWorkspace(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int32) ([]generated.Project, error) {
	if err := s.ensureWorkspaceMember(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return s.Q.ListProjectsForWorkspaceMember(ctx, generated.ListProjectsForWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Limit:       limit,
		Offset:      offset,
	})
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
