package workspace

import (
	"encoding/json"
	"net/http"

	"devpulse-backend/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Svc Service
}

type createWorkspaceRequest struct {
	Name string `json:"name"`
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		authWriteError(w)
		return
	}
	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "invalid_json", "message": "invalid json body"}})
		return
	}
	ws, err := h.Svc.Create(r.Context(), uid, req.Name)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ws)
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		authWriteError(w)
		return
	}
	items, err := h.Svc.ListByUser(r.Context(), uid, 50, 0)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		authWriteError(w)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "invalid_id", "message": "invalid workspace id"}})
		return
	}
	ws, err := h.Svc.GetByID(r.Context(), uid, id)
	if err != nil {
		writeSvcError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ws)
}

func writeSvcError(w http.ResponseWriter, err error) {
	switch err {
	case ErrInvalidName:
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "invalid_name", "message": "name is required"}})
	case ErrNotFound:
		writeJSON(w, http.StatusNotFound, map[string]any{"error": map[string]string{"code": "not_found", "message": "workspace not found"}})
	case ErrUnauthenticated:
		authWriteError(w)
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]string{"code": "internal", "message": "internal server error"}})
	}
}

func authWriteError(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, map[string]any{"error": map[string]string{"code": "unauthorized", "message": "unauthorized"}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

