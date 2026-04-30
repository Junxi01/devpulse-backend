package auth

import (
	"errors"
	"net/http"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrPasswordTooShort = errors.New("password too short")
	ErrEmailTaken      = errors.New("email already exists")
	ErrUnavailable     = errors.New("service unavailable")
)

type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string { return e.Message }

func toAPIError(err error) APIError {
	switch {
	case errors.Is(err, ErrInvalidEmail):
		return APIError{Status: http.StatusBadRequest, Code: "invalid_email", Message: "email is invalid"}
	case errors.Is(err, ErrPasswordTooShort):
		return APIError{Status: http.StatusBadRequest, Code: "invalid_password", Message: "password must be at least 8 characters"}
	case errors.Is(err, ErrEmailTaken):
		return APIError{Status: http.StatusConflict, Code: "email_taken", Message: "email already registered"}
	case errors.Is(err, ErrUnavailable):
		return APIError{Status: http.StatusServiceUnavailable, Code: "unavailable", Message: "service temporarily unavailable"}
	default:
		return APIError{Status: http.StatusInternalServerError, Code: "internal", Message: "internal server error"}
	}
}

