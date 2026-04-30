package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeRegistrar struct {
	fn func(email, password, name string) (RegisteredUser, error)
}

func (f fakeRegistrar) Register(_ context.Context, email, password, name string) (RegisteredUser, error) {
	return f.fn(email, password, name)
}

func TestHandler_Register_Success(t *testing.T) {
	h := Handler{Service: fakeRegistrar{
		fn: func(email, password, name string) (RegisteredUser, error) {
			return RegisteredUser{ID: "u1", Email: email, Name: name}, nil
		},
	}}

	body := `{"email":"a@b.com","password":"password123","name":"n"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d. body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte("password_hash")) {
		t.Fatalf("response must not include password_hash: %s", w.Body.String())
	}
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
	h := Handler{Service: fakeRegistrar{fn: func(string, string, string) (RegisteredUser, error) { return RegisteredUser{}, nil }}}
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("{"))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] == nil {
		t.Fatalf("expected error response, got %s", w.Body.String())
	}
}

func TestHandler_Register_Conflict(t *testing.T) {
	h := Handler{Service: fakeRegistrar{
		fn: func(email, password, name string) (RegisteredUser, error) {
			return RegisteredUser{}, ErrEmailTaken
		},
	}}

	body := `{"email":"a@b.com","password":"password123","name":"n"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d. body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %s", w.Body.String())
	}
	if errObj["code"] != "email_taken" {
		t.Fatalf("code = %v, want %v", errObj["code"], "email_taken")
	}
}

func TestHandler_Register_InternalError(t *testing.T) {
	h := Handler{Service: fakeRegistrar{
		fn: func(email, password, name string) (RegisteredUser, error) {
			return RegisteredUser{}, errors.New("boom")
		},
	}}

	body := `{"email":"a@b.com","password":"password123","name":"n"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d. body=%s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
}

