package mock

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"devpulse-backend/internal/db"
	"devpulse-backend/internal/db/generated"

	"github.com/jackc/pgx/v5"
)

func TestSplitInserted(t *testing.T) {
	t.Parallel()
	ins, skip, err := splitInserted(nil)
	if !ins || skip || err != nil {
		t.Fatalf("want inserted=true for nil err, got ins=%v skip=%v err=%v", ins, skip, err)
	}
	ins, skip, err = splitInserted(pgx.ErrNoRows)
	if ins || !skip || err != nil {
		t.Fatalf("want skipped=true for ErrNoRows, got ins=%v skip=%v err=%v", ins, skip, err)
	}
	ins, skip, err = splitInserted(errors.New("boom"))
	if ins || skip || err == nil {
		t.Fatalf("want err for other errors, got ins=%v skip=%v err=%v", ins, skip, err)
	}
}

func TestDecodeSeedWebhookFiles(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("..", "..", "..", "seed", "github_events")
	for _, name := range []string{"push.json", "pull_request_opened.json", "issues_opened.json"} {
		path := filepath.Join(dir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Skipf("seed fixtures not found at %s (%v)", path, err)
		}
		var env WebhookEnvelope
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if env.DeliveryID == "" || env.EventType == "" || env.OccurredAt.IsZero() {
			t.Fatalf("%s: incomplete envelope %+v", name, env)
		}
	}
}

func TestImportFromDirectory_Idempotent(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	database, err := db.Open(ctx, dsn)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	defer database.Pool.Close()

	repo, err := database.Queries.GetRepositoryByProviderFullName(ctx, generated.GetRepositoryByProviderFullNameParams{
		Provider: DemoRepositoryProvider,
		FullName: DemoRepositoryFullName,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			t.Skip("demo repository missing; run make migrate-up && make seed-demo")
		}
		t.Fatalf("lookup demo repo: %v", err)
	}

	for _, stmt := range []string{
		`DELETE FROM commits WHERE repository_id = $1`,
		`DELETE FROM pull_requests WHERE repository_id = $1`,
		`DELETE FROM issues WHERE repository_id = $1`,
		`DELETE FROM repository_events WHERE repository_id = $1`,
	} {
		if _, err := database.Pool.Exec(ctx, stmt, repo.ID); err != nil {
			t.Fatalf("cleanup activity rows: %v", err)
		}
	}

	dir := filepath.Join("..", "..", "..", "seed", "github_events")
	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}

	st1, err := ImportFromDirectory(ctx, database.Pool, database.Queries, repo.ID, absDir)
	if err != nil {
		t.Fatalf("first import: %v", err)
	}
	if st1.RepositoryEventsInserted != 3 {
		t.Fatalf("first run: want 3 repository_events inserted, got %+v", st1)
	}
	if st1.CommitsInserted != 1 || st1.PullRequestsInserted != 1 || st1.IssuesInserted != 1 {
		t.Fatalf("first run domain rows: %+v", st1)
	}

	st2, err := ImportFromDirectory(ctx, database.Pool, database.Queries, repo.ID, absDir)
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if st2.RepositoryEventsInserted != 0 || st2.RepositoryEventsSkipped != 3 {
		t.Fatalf("second run should skip all events: %+v", st2)
	}
	if st2.PullRequestsInserted != 0 || st2.IssuesInserted != 0 || st2.CommitsInserted != 0 {
		t.Fatalf("second run should not insert domain rows: %+v", st2)
	}
	if st2.PullRequestsSkipped != 1 || st2.IssuesSkipped != 1 || st2.CommitsSkipped != 1 {
		t.Fatalf("second run should skip each domain row once: %+v", st2)
	}
}
