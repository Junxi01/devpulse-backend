package project

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"devpulse-backend/internal/auth"
	"devpulse-backend/internal/db/generated"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func reqWithChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

type forbiddenProjectSvc struct{}

func (forbiddenProjectSvc) CreateForWorkspace(ctx context.Context, userID, workspaceUUID uuid.UUID, name, description string) (generated.Project, error) {
	return generated.Project{}, ErrForbidden
}

func (forbiddenProjectSvc) ListForWorkspace(ctx context.Context, userID, workspaceUUID uuid.UUID, limit, offset int32) ([]generated.Project, error) {
	return nil, ErrForbidden
}

func TestHandler_CreateForWorkspace_UnauthorizedWithoutContextUser(t *testing.T) {
	h := Handler{}
	req := httptest.NewRequest(http.MethodPost, "/workspaces/"+uuid.New().String()+"/projects", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.CreateForWorkspace(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateForWorkspace_Forbidden(t *testing.T) {
	h := Handler{Svc: forbiddenProjectSvc{}}
	uid := uuid.New()
	ws := uuid.New()
	body := bytes.NewBufferString(`{"name":"demo"}`)
	req := httptest.NewRequest(http.MethodPost, "/workspaces/"+ws.String()+"/projects", body)
	req.Header.Set("Content-Type", "application/json")
	req = reqWithChiParam(req, "workspaceID", ws.String())
	req = req.WithContext(auth.ContextWithUserID(req.Context(), uid))
	w := httptest.NewRecorder()
	h.CreateForWorkspace(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_ListForWorkspace_UnauthorizedWithoutContextUser(t *testing.T) {
	h := Handler{}
	req := httptest.NewRequest(http.MethodGet, "/workspaces/"+uuid.New().String()+"/projects", nil)
	w := httptest.NewRecorder()
	h.ListForWorkspace(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_ListForWorkspace_Forbidden(t *testing.T) {
	h := Handler{Svc: forbiddenProjectSvc{}}
	uid := uuid.New()
	ws := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/workspaces/"+ws.String()+"/projects", nil)
	req = reqWithChiParam(req, "workspaceID", ws.String())
	req = req.WithContext(auth.ContextWithUserID(req.Context(), uid))
	w := httptest.NewRecorder()
	h.ListForWorkspace(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
