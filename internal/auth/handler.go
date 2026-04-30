package auth

import (
	"context"
	"encoding/json"
	"net/http"
)

type Registrar interface {
	Register(ctx context.Context, email, password, name string) (RegisteredUser, error)
}

type Handler struct {
	Service Registrar
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

	if h.Service == nil {
		writeError(w, toAPIError(ErrUnavailable))
		return
	}

	u, err := h.Service.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		writeError(w, toAPIError(err))
		return
	}

	writeJSON(w, http.StatusCreated, u)
}

func writeError(w http.ResponseWriter, apiErr APIError) {
	writeJSON(w, apiErr.Status, errorResponse{Error: APIError{Code: apiErr.Code, Message: apiErr.Message}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

