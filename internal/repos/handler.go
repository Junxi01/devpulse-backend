package repos

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"devpulse-backend/internal/auth"
	"devpulse-backend/internal/db/generated"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SvcAPI interface {
	CreateForProject(ctx context.Context, userID, projectID uuid.UUID, provider, owner, name, fullName, externalID, defaultBranch string) (generated.Repository, error)
	ListForProject(ctx context.Context, userID, projectID uuid.UUID, limit, offset int32) ([]generated.Repository, error)
}

type Handler struct {
	Svc SvcAPI
}

type createRepositoryRequest struct {
	Provider      string `json:"provider"`
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	ExternalID    string `json:"external_id"`
	DefaultBranch string `json:"default_branch"`
}

func (h Handler) CreateForProject(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}
	projectStr := chi.URLParam(r, "projectID")
	projectID, err := uuid.Parse(projectStr)
	if err != nil {
		writeBadRequest(w, "invalid_project_id", "invalid project id")
		return
	}
	var req createRepositoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "invalid_json", "message": "invalid json body"}})
		return
	}
	repo, err := h.Svc.CreateForProject(r.Context(), uid, projectID, req.Provider, req.Owner, req.Name, req.FullName, req.ExternalID, req.DefaultBranch)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, repo)
}

func (h Handler) ListForProject(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}
	projectStr := chi.URLParam(r, "projectID")
	projectID, err := uuid.Parse(projectStr)
	if err != nil {
		writeBadRequest(w, "invalid_project_id", "invalid project id")
		return
	}
	limit := int32(50)
	offset := int32(0)
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
	items, err := h.Svc.ListForProject(r.Context(), uid, projectID, limit, offset)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	if items == nil {
		items = []generated.Repository{}
	}
	writeJSON(w, http.StatusOK, items)
}

func writeSvcError(w http.ResponseWriter, err error) {
	switch err {
	case ErrInvalidInput:
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "invalid_input", "message": "provider, owner, name, full_name and external_id are required"}})
	case ErrForbidden:
		writeJSON(w, http.StatusForbidden, map[string]any{"error": map[string]string{"code": "forbidden", "message": "project not accessible in this workspace"}})
	case ErrConflict:
		writeJSON(w, http.StatusConflict, map[string]any{"error": map[string]string{"code": "conflict", "message": "repository already linked"}})
	case ErrUnauthenticated:
		writeUnauthorized(w)
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]string{"code": "internal", "message": "internal server error"}})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, map[string]any{"error": map[string]string{"code": "unauthorized", "message": "unauthorized"}})
}

func writeBadRequest(w http.ResponseWriter, code, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
