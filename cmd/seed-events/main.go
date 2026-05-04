// Command seed-events imports mock GitHub webhook JSON into the demo repository (local only).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devpulse-backend/internal/config"
	"devpulse-backend/internal/db"
	"devpulse-backend/internal/db/generated"
	ghmock "devpulse-backend/internal/github/mock"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func main() {
	if err := run(context.Background()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "seed-events: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	repoFlag := flag.String("repo", "", "target repository UUID (default: demo repo from seed-demo)")
	dirFlag := flag.String("dir", "", "directory with *.json mock deliveries (default: ./seed/github_events)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if !seedAllowed(cfg) {
		return fmt.Errorf("refusing to import: set APP_MODE=demo (default) or SEED_DEMO=1")
	}

	dir := strings.TrimSpace(*dirFlag)
	if dir == "" {
		dir = filepath.Join("seed", "github_events")
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	database, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer database.Pool.Close()

	var repoID uuid.UUID
	if s := strings.TrimSpace(*repoFlag); s != "" {
		repoID, err = uuid.Parse(s)
		if err != nil {
			return fmt.Errorf("parse -repo: %w", err)
		}
	} else {
		r, err := database.Queries.GetRepositoryByProviderFullName(ctx, generated.GetRepositoryByProviderFullNameParams{
			Provider: ghmock.DemoRepositoryProvider,
			FullName: ghmock.DemoRepositoryFullName,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("demo repository not found (expected %s/%s); run make seed-demo first", ghmock.DemoRepositoryProvider, ghmock.DemoRepositoryFullName)
			}
			return err
		}
		repoID = r.ID
	}

	stats, err := ghmock.ImportFromDirectory(ctx, database.Pool, database.Queries, repoID, absDir)
	if err != nil {
		return err
	}

	fmt.Printf("seed-events: ok repo_id=%s dir=%s\n", repoID, absDir)
	fmt.Printf("  repository_events: inserted=%d skipped=%d\n", stats.RepositoryEventsInserted, stats.RepositoryEventsSkipped)
	fmt.Printf("  pull_requests:    inserted=%d skipped=%d\n", stats.PullRequestsInserted, stats.PullRequestsSkipped)
	fmt.Printf("  issues:           inserted=%d skipped=%d\n", stats.IssuesInserted, stats.IssuesSkipped)
	fmt.Printf("  commits:          inserted=%d skipped=%d\n", stats.CommitsInserted, stats.CommitsSkipped)
	return nil
}

func seedAllowed(cfg config.Config) bool {
	if os.Getenv("SEED_DEMO") == "1" {
		return true
	}
	return cfg.AppMode == "demo"
}
