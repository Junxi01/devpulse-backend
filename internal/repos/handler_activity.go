package repos

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"devpulse-backend/internal/auth"
	"devpulse-backend/internal/db/generated"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ptrPgText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

type pullRequestDTO struct {
	ID           uuid.UUID `json:"id"`
	RepositoryID uuid.UUID `json:"repository_id"`
	Number       int32     `json:"number"`
	Title        string    `json:"title"`
	State        string    `json:"state"`
	Author       string    `json:"author"`
	BaseBranch   string    `json:"base_branch"`
	HeadBranch   string    `json:"head_branch"`
	ChangedFiles int32     `json:"changed_files"`
	Additions    int32     `json:"additions"`
	Deletions    int32     `json:"deletions"`
	RiskLevel    *string   `json:"risk_level"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
}

func pullRequestsToDTO(rows []generated.PullRequest) []pullRequestDTO {
	out := make([]pullRequestDTO, len(rows))
	for i, p := range rows {
		out[i] = pullRequestDTO{
			ID:           p.ID,
			RepositoryID: p.RepositoryID,
			Number:       p.Number,
			Title:        p.Title,
			State:        p.State,
			Author:       p.Author,
			BaseBranch:   p.BaseBranch,
			HeadBranch:   p.HeadBranch,
			ChangedFiles: p.ChangedFiles,
			Additions:    p.Additions,
			Deletions:    p.Deletions,
			RiskLevel:    ptrPgText(p.RiskLevel),
			CreatedAt:    p.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    p.UpdatedAt.Format(time.RFC3339Nano),
		}
	}
	return out
}

type issueDTO struct {
	ID           uuid.UUID       `json:"id"`
	RepositoryID uuid.UUID       `json:"repository_id"`
	Number       int32           `json:"number"`
	Title        string          `json:"title"`
	State        string          `json:"state"`
	Author       string          `json:"author"`
	Labels       json.RawMessage `json:"labels"`
	Priority     *string         `json:"priority"`
	Category     *string         `json:"category"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
}

func issuesToDTO(rows []generated.Issue) []issueDTO {
	out := make([]issueDTO, len(rows))
	for i, x := range rows {
		labels := x.Labels
		if labels == nil {
			labels = json.RawMessage(`[]`)
		}
		out[i] = issueDTO{
			ID:           x.ID,
			RepositoryID: x.RepositoryID,
			Number:       x.Number,
			Title:        x.Title,
			State:        x.State,
			Author:       x.Author,
			Labels:       labels,
			Priority:     ptrPgText(x.Priority),
			Category:     ptrPgText(x.Category),
			CreatedAt:    x.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    x.UpdatedAt.Format(time.RFC3339Nano),
		}
	}
	return out
}

func (h Handler) ListRepositoryEvents(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}
	repoID, err := uuid.Parse(chi.URLParam(r, "repoID"))
	if err != nil {
		writeBadRequest(w, "invalid_repository_id", "invalid repository id")
		return
	}
	limit, offset := parseLimitOffset(r)
	items, err := h.Svc.ListEventsForRepository(r.Context(), uid, repoID, limit, offset)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	if items == nil {
		items = []generated.RepositoryEvent{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h Handler) ListRepositoryPullRequests(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}
	repoID, err := uuid.Parse(chi.URLParam(r, "repoID"))
	if err != nil {
		writeBadRequest(w, "invalid_repository_id", "invalid repository id")
		return
	}
	limit, offset := parseLimitOffset(r)
	items, err := h.Svc.ListPullRequestsForRepository(r.Context(), uid, repoID, limit, offset)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	if items == nil {
		items = []generated.PullRequest{}
	}
	writeJSON(w, http.StatusOK, pullRequestsToDTO(items))
}

func (h Handler) ListRepositoryIssues(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}
	repoID, err := uuid.Parse(chi.URLParam(r, "repoID"))
	if err != nil {
		writeBadRequest(w, "invalid_repository_id", "invalid repository id")
		return
	}
	limit, offset := parseLimitOffset(r)
	items, err := h.Svc.ListIssuesForRepository(r.Context(), uid, repoID, limit, offset)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	if items == nil {
		items = []generated.Issue{}
	}
	writeJSON(w, http.StatusOK, issuesToDTO(items))
}

func parseLimitOffset(r *http.Request) (limit, offset int32) {
	limit = 50
	offset = 0
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if v, err := strconv.ParseInt(ls, 10, 32); err == nil && v > 0 {
			limit = int32(v)
		}
	}
	if os := r.URL.Query().Get("offset"); os != "" {
		if v, err := strconv.ParseInt(os, 10, 32); err == nil && v >= 0 {
			offset = int32(v)
		}
	}
	return limit, offset
}
