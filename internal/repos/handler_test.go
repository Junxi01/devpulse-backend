package repos

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

type forbiddenReposSvc struct{}

func (forbiddenReposSvc) CreateForProject(ctx context.Context, userID, projectID uuid.UUID, provider, owner, name, fullName, externalID, defaultBranch string) (generated.Repository, error) {
	return generated.Repository{}, ErrForbidden
}

func (forbiddenReposSvc) ListForProject(ctx context.Context, userID, projectID uuid.UUID, limit, offset int32) ([]generated.Repository, error) {
	return nil, ErrForbidden
}

func TestHandler_CreateForProject_UnauthorizedWithoutContextUser(t *testing.T) {
	h := Handler{}
	body := `{"provider":"github","owner":"a","name":"b","full_name":"a/b","external_id":"1"}`
	req := httptest.NewRequest(http.MethodPost, "/projects/"+uuid.New().String()+"/repositories", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.CreateForProject(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateForProject_Forbidden(t *testing.T) {
	h := Handler{Svc: forbiddenReposSvc{}}
	uid := uuid.New()
	pid := uuid.New()
	body := `{"provider":"github","owner":"a","name":"b","full_name":"a/b","external_id":"1"}`
	req := httptest.NewRequest(http.MethodPost, "/projects/"+pid.String()+"/repositories", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = reqWithChiParam(req, "projectID", pid.String())
	req = req.WithContext(auth.ContextWithUserID(req.Context(), uid))
	w := httptest.NewRecorder()
	h.CreateForProject(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_ListForProject_UnauthorizedWithoutContextUser(t *testing.T) {
	h := Handler{}
	req := httptest.NewRequest(http.MethodGet, "/projects/"+uuid.New().String()+"/repositories", nil)
	w := httptest.NewRecorder()
	h.ListForProject(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_ListForProject_Forbidden(t *testing.T) {
	h := Handler{Svc: forbiddenReposSvc{}}
	uid := uuid.New()
	pid := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/projects/"+pid.String()+"/repositories", nil)
	req = reqWithChiParam(req, "projectID", pid.String())
	req = req.WithContext(auth.ContextWithUserID(req.Context(), uid))
	w := httptest.NewRecorder()
	h.ListForProject(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
