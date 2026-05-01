package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type fakeLoginSvc struct {
	fn func(email, password string) (LoginResponse, error)
}

func (f fakeLoginSvc) Login(_ context.Context, email, password string) (LoginResponse, error) {
	return f.fn(email, password)
}

func TestHandler_Login_Success(t *testing.T) {
	h := Handler{LoginSvc: fakeLoginSvc{
		fn: func(email, password string) (LoginResponse, error) {
			return LoginResponse{
				AccessToken: "tok",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				User:        RegisteredUser{ID: "u1", Email: email, Name: "n"},
			}, nil
		},
	}}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"a@b.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp LoginResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.AccessToken == "" || resp.TokenType != "Bearer" || resp.ExpiresIn == 0 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestMiddleware_RequireAuth(t *testing.T) {
	secret := "s3cr3t"
	mw := Middleware{JWTSecret: secret}

	claims := jwt.MapClaims{
		"user_id": "00000000-0000-0000-0000-000000000001",
		"email":   "a@b.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte(secret))

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	w := httptest.NewRecorder()

	mw.RequireAuth(next).ServeHTTP(w, req)
	if !called || w.Code != http.StatusOK {
		t.Fatalf("called=%v status=%d", called, w.Code)
	}
}

func TestMiddleware_MissingToken(t *testing.T) {
	mw := Middleware{JWTSecret: "x"}
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

