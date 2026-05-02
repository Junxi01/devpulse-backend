package project

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
	CreateForWorkspace(ctx context.Context, userID, workspaceID uuid.UUID, name, description string) (generated.Project, error)
	ListForWorkspace(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int32) ([]generated.Project, error)
}

type Handler struct {
	Svc SvcAPI
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h Handler) CreateForWorkspace(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}
	wsStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(wsStr)
	if err != nil {
		writeBadRequest(w, "invalid_workspace_id", "invalid workspace id")
		return
	}
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "invalid_json", "message": "invalid json body"}})
		return
	}
	p, err := h.Svc.CreateForWorkspace(r.Context(), uid, workspaceID, req.Name, req.Description)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h Handler) ListForWorkspace(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeUnauthorized(w)
		return
	}
	wsStr := chi.URLParam(r, "workspaceID")
	workspaceID, err := uuid.Parse(wsStr)
	if err != nil {
		writeBadRequest(w, "invalid_workspace_id", "invalid workspace id")
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
	items, err := h.Svc.ListForWorkspace(r.Context(), uid, workspaceID, limit, offset)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	if items == nil {
		items = []generated.Project{}
	}
	writeJSON(w, http.StatusOK, items)
}

func writeSvcError(w http.ResponseWriter, err error) {
	switch err {
	case ErrInvalidName:
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "invalid_name", "message": "name is required"}})
	case ErrForbidden:
		writeJSON(w, http.StatusForbidden, map[string]any{"error": map[string]string{"code": "forbidden", "message": "not a workspace member"}})
	case ErrConflict:
		writeJSON(w, http.StatusConflict, map[string]any{"error": map[string]string{"code": "conflict", "message": "project already exists"}})
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
