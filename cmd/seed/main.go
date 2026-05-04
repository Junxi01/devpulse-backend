// Command seed loads idempotent demo data for local development only.
// Credentials exist only in this command (not in API handlers). Run via: make seed-demo
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"devpulse-backend/internal/config"
	"devpulse-backend/internal/db"
	"devpulse-backend/internal/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

const (
	demoEmail          = "demo@devpulse.local"
	demoPassword       = "demo123456"
	demoUserName       = "Demo User"
	demoWorkspaceName  = "Demo Workspace"
	demoProjectName    = "Demo Project"
	demoRepoProvider   = "github"
	demoRepoOwner      = "devpulse"
	demoRepoName       = "demo-backend"
	demoRepoFullName   = "devpulse/demo-backend"
	demoRepoExternalID = "1001"
)

func main() {
	if err := run(context.Background()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("seed: demo data ready")
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if !seedAllowed(cfg) {
		return errors.New("refusing to seed: set APP_MODE=demo (default) or SEED_DEMO=1 for a one-off run")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	database, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer database.Pool.Close()

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
INSERT INTO users (email, password_hash, name)
VALUES ($1, $2, $3)
ON CONFLICT (email) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  name = EXCLUDED.name,
  updated_at = now()
`, demoEmail, string(hash), demoUserName); err != nil {
		return fmt.Errorf("upsert demo user: %w", err)
	}

	q := database.Queries.WithTx(tx)

	u, err := q.GetUserByEmail(ctx, demoEmail)
	if err != nil {
		return fmt.Errorf("load demo user: %w", err)
	}

	ws, err := ensureWorkspace(ctx, q, u.ID)
	if err != nil {
		return err
	}

	proj, err := ensureProject(ctx, q, u.ID, ws.ID)
	if err != nil {
		return err
	}

	if err := ensureRepository(ctx, q, u.ID, proj.ID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func seedAllowed(cfg config.Config) bool {
	if strings.TrimSpace(os.Getenv("SEED_DEMO")) == "1" {
		return true
	}
	return cfg.AppMode == "demo"
}

func ensureWorkspace(ctx context.Context, q *generated.Queries, userID uuid.UUID) (generated.Workspace, error) {
	list, err := q.ListWorkspacesByUser(ctx, generated.ListWorkspacesByUserParams{
		UserID: userID,
		Limit:  200,
		Offset: 0,
	})
	if err != nil {
		return generated.Workspace{}, err
	}
	for _, w := range list {
		if w.Name == demoWorkspaceName {
			return w, nil
		}
	}
	ws, err := q.CreateWorkspace(ctx, generated.CreateWorkspaceParams{
		Name:    demoWorkspaceName,
		OwnerID: userID,
	})
	if err != nil {
		return generated.Workspace{}, fmt.Errorf("create workspace: %w", err)
	}
	if err := q.AddWorkspaceMember(ctx, generated.AddWorkspaceMemberParams{
		WorkspaceID: ws.ID,
		UserID:      userID,
		Role:        "owner",
	}); err != nil {
		return generated.Workspace{}, fmt.Errorf("workspace membership: %w", err)
	}
	return ws, nil
}

func ensureProject(ctx context.Context, q *generated.Queries, userID, workspaceID uuid.UUID) (generated.Project, error) {
	list, err := q.ListProjectsForWorkspaceMember(ctx, generated.ListProjectsForWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Limit:       200,
		Offset:      0,
	})
	if err != nil {
		return generated.Project{}, err
	}
	for _, p := range list {
		if p.Name == demoProjectName {
			return p, nil
		}
	}
	return q.CreateProject(ctx, generated.CreateProjectParams{
		WorkspaceID: workspaceID,
		Name:        demoProjectName,
		Description: "Seeded demo project for local development.",
	})
}

func ensureRepository(ctx context.Context, q *generated.Queries, userID, projectID uuid.UUID) error {
	list, err := q.ListRepositoriesForWorkspaceMember(ctx, generated.ListRepositoriesForWorkspaceMemberParams{
		ProjectID: projectID,
		UserID:    userID,
		Limit:     200,
		Offset:    0,
	})
	if err != nil {
		return err
	}
	for _, r := range list {
		if r.Provider == demoRepoProvider && r.ExternalID == demoRepoExternalID {
			return nil
		}
	}
	_, err = q.CreateRepository(ctx, generated.CreateRepositoryParams{
		ProjectID:     projectID,
		Provider:      demoRepoProvider,
		Owner:         demoRepoOwner,
		Name:          demoRepoName,
		FullName:      demoRepoFullName,
		ExternalID:    demoRepoExternalID,
		DefaultBranch: "main",
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil
		}
		return fmt.Errorf("create repository: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
