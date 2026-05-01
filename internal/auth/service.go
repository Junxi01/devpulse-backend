package auth

import (
	"context"
	"net/mail"
	"time"
	"strings"

	"devpulse-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Users     repository.UserRepository
	JWTSecret string
	AccessTTL time.Duration
}

type RegisteredUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type LoginResponse struct {
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"`
	ExpiresIn   int64          `json:"expires_in"`
	User        RegisteredUser `json:"user"`
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

func (s Service) Login(ctx context.Context, email, password string) (LoginResponse, error) {
	if s.Users == nil || s.JWTSecret == "" {
		return LoginResponse{}, ErrUnavailable
	}
	email = strings.TrimSpace(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return LoginResponse{}, ErrUnauthorized
	}
	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return LoginResponse{}, ErrUnauthorized
		}
		return LoginResponse{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return LoginResponse{}, ErrUnauthorized
	}
	if s.AccessTTL <= 0 {
		s.AccessTTL = time.Hour
	}
	now := time.Now()
	exp := now.Add(s.AccessTTL)

	claims := jwt.MapClaims{
		"user_id": u.ID.String(),
		"email":   u.Email,
		"exp":     exp.Unix(),
		"iat":     now.Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(s.JWTSecret))
	if err != nil {
		return LoginResponse{}, err
	}

	user := RegisteredUser{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return LoginResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.AccessTTL.Seconds()),
		User:        user,
	}, nil
}

func (s Service) Me(ctx context.Context, userID uuid.UUID) (RegisteredUser, error) {
	if s.Users == nil {
		return RegisteredUser{}, ErrUnavailable
	}
	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return RegisteredUser{}, ErrUnauthorized
		}
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

