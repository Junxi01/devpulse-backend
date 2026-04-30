package auth

import (
	"context"
	"net/mail"
	"strings"

	"devpulse-backend/internal/repository"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Users repository.UserRepository
}

type RegisteredUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s Service) Register(ctx context.Context, email, password, name string) (RegisteredUser, error) {
	if s.Users == nil {
		return RegisteredUser{}, ErrUnavailable
	}

	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)

	if _, err := mail.ParseAddress(email); err != nil {
		return RegisteredUser{}, ErrInvalidEmail
	}
	if len(password) < 8 {
		return RegisteredUser{}, ErrPasswordTooShort
	}

	// Check uniqueness.
	_, err := s.Users.GetByEmail(ctx, email)
	if err == nil {
		return RegisteredUser{}, ErrEmailTaken
	}
	if err != nil && err != pgx.ErrNoRows {
		return RegisteredUser{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return RegisteredUser{}, err
	}

	u, err := s.Users.Create(ctx, email, string(hash), name)
	if err != nil {
		return RegisteredUser{}, err
	}

	return RegisteredUser{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

