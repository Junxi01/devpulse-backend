package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devpulse-backend/internal/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// DemoRepositoryProvider and DemoRepositoryFullName match data created by `make seed-demo`.
	DemoRepositoryProvider = "github"
	DemoRepositoryFullName = "devpulse/demo-backend"
)

// Stats counts inserted vs skipped rows for idempotent operations (skipped = conflict / already present).
type Stats struct {
	RepositoryEventsInserted int
	RepositoryEventsSkipped  int
	PullRequestsInserted     int
	PullRequestsSkipped      int
	IssuesInserted           int
	IssuesSkipped            int
	CommitsInserted          int
	CommitsSkipped           int
}

// Add accumulates another Stats into s.
func (s *Stats) Add(o Stats) {
	s.RepositoryEventsInserted += o.RepositoryEventsInserted
	s.RepositoryEventsSkipped += o.RepositoryEventsSkipped
	s.PullRequestsInserted += o.PullRequestsInserted
	s.PullRequestsSkipped += o.PullRequestsSkipped
	s.IssuesInserted += o.IssuesInserted
	s.IssuesSkipped += o.IssuesSkipped
	s.CommitsInserted += o.CommitsInserted
	s.CommitsSkipped += o.CommitsSkipped
}

// ImportFromDirectory runs all `*.json` files in dir (sorted by name) in one transaction.
func ImportFromDirectory(ctx context.Context, pool *pgxpool.Pool, q *generated.Queries, repositoryID uuid.UUID, dir string) (Stats, error) {
	if q == nil {
		return Stats{}, errors.New("queries is nil")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Stats{}, fmt.Errorf("read seed dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return Stats{}, fmt.Errorf("no .json files in %s", dir)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return Stats{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := q.WithTx(tx)
	var total Stats
	for _, name := range names {
		path := filepath.Join(dir, name)
		st, err := importFile(ctx, qtx, repositoryID, path)
		if err != nil {
			return total, fmt.Errorf("%s: %w", name, err)
		}
		total.Add(st)
	}
	if err := tx.Commit(ctx); err != nil {
		return total, err
	}
	return total, nil
}

func importFile(ctx context.Context, q *generated.Queries, repositoryID uuid.UUID, path string) (Stats, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Stats{}, err
	}
	return ImportJSON(ctx, q, repositoryID, raw)
}

// ImportJSON parses a single mock webhook JSON document and applies idempotent inserts.
func ImportJSON(ctx context.Context, q *generated.Queries, repositoryID uuid.UUID, raw []byte) (Stats, error) {
	var env WebhookEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return Stats{}, fmt.Errorf("decode envelope: %w", err)
	}
	if strings.TrimSpace(env.DeliveryID) == "" {
		return Stats{}, errors.New("delivery_id is required")
	}
	if strings.TrimSpace(env.EventType) == "" {
		return Stats{}, errors.New("event_type is required")
	}
	if env.OccurredAt.IsZero() {
		return Stats{}, errors.New("occurred_at is required")
	}

	var st Stats
	ins, _, err := tryInsertEvent(ctx, q, generated.InsertRepositoryEventIdempotentParams{
		RepositoryID: repositoryID,
		EventType:    env.EventType,
		ExternalID:   env.DeliveryID,
		Payload:      ensurePayloadRaw(env.Payload),
		OccurredAt:   env.OccurredAt,
	})
	if err != nil {
		return st, err
	}
	if ins {
		st.RepositoryEventsInserted++
	} else {
		st.RepositoryEventsSkipped++
	}

	et := strings.ToLower(env.EventType)
	switch et {
	case "push":
		s2, err := importPush(ctx, q, repositoryID, env)
		if err != nil {
			return st, err
		}
		st.Add(s2)
		return st, nil
	case "pull_request", "pull_request_opened":
		s2, err := importPullRequest(ctx, q, repositoryID, env)
		if err != nil {
			return st, err
		}
		st.Add(s2)
		return st, nil
	case "issues", "issue", "issues_opened":
		s2, err := importIssue(ctx, q, repositoryID, env)
		if err != nil {
			return st, err
		}
		st.Add(s2)
		return st, nil
	default:
		return st, fmt.Errorf("unsupported event_type %q", env.EventType)
	}
}

func tryInsertEvent(ctx context.Context, q *generated.Queries, arg generated.InsertRepositoryEventIdempotentParams) (inserted, skipped bool, err error) {
	_, err = q.InsertRepositoryEventIdempotent(ctx, arg)
	return splitInserted(err)
}

func splitInserted(err error) (inserted, skipped bool, errOut error) {
	if err == nil {
		return true, false, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, true, nil
	}
	return false, false, err
}

func importPush(ctx context.Context, q *generated.Queries, repositoryID uuid.UUID, env WebhookEnvelope) (Stats, error) {
	var st Stats
	var p pushPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return st, fmt.Errorf("decode push payload: %w", err)
	}
	for _, c := range p.Commits {
		if strings.TrimSpace(c.ID) == "" {
			continue
		}
		at := parseTime(c.Timestamp, env.OccurredAt)
		_, err := q.InsertCommitIdempotent(ctx, generated.InsertCommitIdempotentParams{
			RepositoryID: repositoryID,
			Sha:          strings.TrimSpace(c.ID),
			Message:      strings.TrimSpace(c.Message),
			Author:       authorName(c, "unknown"),
			CommittedAt:  at,
		})
		ins, skip, err := splitInserted(err)
		if err != nil {
			return st, err
		}
		if ins {
			st.CommitsInserted++
		}
		if skip {
			st.CommitsSkipped++
		}
	}
	return st, nil
}

func importPullRequest(ctx context.Context, q *generated.Queries, repositoryID uuid.UUID, env WebhookEnvelope) (Stats, error) {
	var st Stats
	var p pullRequestPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return st, fmt.Errorf("decode pull_request payload: %w", err)
	}
	pr := p.PullRequest
	if pr.Number <= 0 {
		return st, errors.New("pull_request.number is required")
	}
	risk := pgtype.Text{Valid: false}
	_, err := q.InsertPullRequestIdempotent(ctx, generated.InsertPullRequestIdempotentParams{
		RepositoryID: repositoryID,
		Number:       int32(pr.Number),
		Title:        pr.Title,
		State:        pr.State,
		Author:       pr.User.Login,
		BaseBranch:   pr.Base.Ref,
		HeadBranch:   pr.Head.Ref,
		ChangedFiles: int32(pr.ChangedFiles),
		Additions:    int32(pr.Additions),
		Deletions:    int32(pr.Deletions),
		RiskLevel:    risk,
	})
	ins, skip, err := splitInserted(err)
	if err != nil {
		return st, err
	}
	if ins {
		st.PullRequestsInserted++
	}
	if skip {
		st.PullRequestsSkipped++
	}
	return st, nil
}

func importIssue(ctx context.Context, q *generated.Queries, repositoryID uuid.UUID, env WebhookEnvelope) (Stats, error) {
	var st Stats
	var p issuesPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return st, fmt.Errorf("decode issues payload: %w", err)
	}
	is := p.Issue
	if is.Number <= 0 {
		return st, errors.New("issue.number is required")
	}
	labelsRaw, err := json.Marshal(is.Labels)
	if err != nil {
		return st, err
	}
	if len(labelsRaw) == 0 {
		labelsRaw = []byte(`[]`)
	}
	prio := pgtype.Text{Valid: false}
	cat := pgtype.Text{Valid: false}
	_, err = q.InsertIssueIdempotent(ctx, generated.InsertIssueIdempotentParams{
		RepositoryID: repositoryID,
		Number:       int32(is.Number),
		Title:        is.Title,
		State:        is.State,
		Author:       is.User.Login,
		Labels:       json.RawMessage(labelsRaw),
		Priority:     prio,
		Category:     cat,
	})
	ins, skip, err := splitInserted(err)
	if err != nil {
		return st, err
	}
	if ins {
		st.IssuesInserted++
	}
	if skip {
		st.IssuesSkipped++
	}
	return st, nil
}
