package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Registrar interface {
	Register(ctx context.Context, email, password, name string) (RegisteredUser, error)
}

type LoginServicer interface {
	Login(ctx context.Context, email, password string) (LoginResponse, error)
}

type MeServicer interface {
	Me(ctx context.Context, userID uuid.UUID) (RegisteredUser, error)
}

type Handler struct {
	Registrar Registrar
	LoginSvc  LoginServicer
	MeSvc     MeServicer
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type errorResponse struct {
	Error APIError `json:"error"`
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, APIError{Status: http.StatusBadRequest, Code: "invalid_json", Message: "invalid json body"})
		return
	}

	if h.Registrar == nil {
		writeError(w, toAPIError(ErrUnavailable))
		return
	}

	u, err := h.Registrar.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		writeError(w, toAPIError(err))
		return
	}

	writeJSON(w, http.StatusCreated, u)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, APIError{Status: http.StatusBadRequest, Code: "invalid_json", Message: "invalid json body"})
		return
	}
	if h.LoginSvc == nil {
		writeError(w, toAPIError(ErrUnavailable))
		return
	}
	resp, err := h.LoginSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, toAPIError(err))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	if h.MeSvc == nil {
		writeError(w, toAPIError(ErrUnavailable))
		return
	}
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, toAPIError(ErrUnauthorized))
		return
	}
	u, err := h.MeSvc.Me(r.Context(), uid)
	if err != nil {
		writeError(w, toAPIError(err))
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func writeError(w http.ResponseWriter, apiErr APIError) {
	writeJSON(w, apiErr.Status, errorResponse{Error: APIError{Code: apiErr.Code, Message: apiErr.Message}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

