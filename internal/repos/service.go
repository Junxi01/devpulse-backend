package repos

import (
	"context"
	"errors"
	"strings"

	"devpulse-backend/internal/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidInput    = errors.New("invalid repository payload")
	ErrConflict        = errors.New("conflict")
)

type Service struct {
	Q *generated.Queries
}

func (s Service) ensureProjectMembership(ctx context.Context, userID, projectID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrUnauthenticated
	}
	if s.Q == nil {
		return errors.New("service unavailable")
	}
	_, err := s.Q.GetProjectForWorkspaceMember(ctx, generated.GetProjectForWorkspaceMemberParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}
	return nil
}

func normalizeRepo(provider, repoOwner, name, fullName, externalID string) (providerOut, repoOwnerOut, nameOut, fullNameOut, externalIDOut string, err error) {
	providerOut = strings.TrimSpace(provider)
	repoOwnerOut = strings.TrimSpace(repoOwner)
	nameOut = strings.TrimSpace(name)
	fullNameOut = strings.TrimSpace(fullName)
	externalIDOut = strings.TrimSpace(externalID)
	if providerOut == "" || repoOwnerOut == "" || nameOut == "" || fullNameOut == "" || externalIDOut == "" {
		return "", "", "", "", "", ErrInvalidInput
	}
	return providerOut, repoOwnerOut, nameOut, fullNameOut, externalIDOut, nil
}

func (s Service) CreateForProject(ctx context.Context, userID, projectID uuid.UUID, provider, repoOwner, name, fullName, externalID, defaultBranch string) (generated.Repository, error) {
	if err := s.ensureProjectMembership(ctx, userID, projectID); err != nil {
		return generated.Repository{}, err
	}
	provider, repoOwner, name, fullName, externalID, err := normalizeRepo(provider, repoOwner, name, fullName, externalID)
	if err != nil {
		return generated.Repository{}, err
	}
	branch := strings.TrimSpace(defaultBranch)
	if branch == "" {
		branch = "main"
	}

	r, err := s.Q.CreateRepository(ctx, generated.CreateRepositoryParams{
		ProjectID:     projectID,
		Provider:      provider,
		Owner:         repoOwner,
		Name:          name,
		FullName:      fullName,
		ExternalID:    externalID,
		DefaultBranch: branch,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return generated.Repository{}, ErrConflict
		}
		return generated.Repository{}, err
	}
	return r, nil
}

func (s Service) ListForProject(ctx context.Context, userID, projectID uuid.UUID, limit, offset int32) ([]generated.Repository, error) {
	if err := s.ensureProjectMembership(ctx, userID, projectID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return s.Q.ListRepositoriesForWorkspaceMember(ctx, generated.ListRepositoriesForWorkspaceMemberParams{
		ProjectID: projectID,
		UserID:    userID,
		Limit:     limit,
		Offset:    offset,
	})
}

func (s Service) ensureRepositoryAccess(ctx context.Context, userID, repositoryID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrUnauthenticated
	}
	if s.Q == nil {
		return errors.New("service unavailable")
	}
	_, err := s.Q.GetRepositoryForWorkspaceMember(ctx, generated.GetRepositoryForWorkspaceMemberParams{
		ID:     repositoryID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}
	return nil
}

func (s Service) ListEventsForRepository(ctx context.Context, userID, repositoryID uuid.UUID, limit, offset int32) ([]generated.RepositoryEvent, error) {
	if err := s.ensureRepositoryAccess(ctx, userID, repositoryID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return s.Q.ListRepositoryEventsForMember(ctx, generated.ListRepositoryEventsForMemberParams{
		RepositoryID: repositoryID,
		UserID:       userID,
		Limit:        limit,
		Offset:       offset,
	})
}

func (s Service) ListPullRequestsForRepository(ctx context.Context, userID, repositoryID uuid.UUID, limit, offset int32) ([]generated.PullRequest, error) {
	if err := s.ensureRepositoryAccess(ctx, userID, repositoryID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return s.Q.ListPullRequestsForMember(ctx, generated.ListPullRequestsForMemberParams{
		RepositoryID: repositoryID,
		UserID:       userID,
		Limit:        limit,
		Offset:       offset,
	})
}

func (s Service) ListIssuesForRepository(ctx context.Context, userID, repositoryID uuid.UUID, limit, offset int32) ([]generated.Issue, error) {
	if err := s.ensureRepositoryAccess(ctx, userID, repositoryID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return s.Q.ListIssuesForMember(ctx, generated.ListIssuesForMemberParams{
		RepositoryID: repositoryID,
		UserID:       userID,
		Limit:        limit,
		Offset:       offset,
	})
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
