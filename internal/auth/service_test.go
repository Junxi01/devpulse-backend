package auth

import (
	"context"
	"testing"

	"devpulse-backend/internal/db/generated"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type fakeUsersRepo struct {
	getByEmail func(ctx context.Context, email string) (generated.User, error)
	create     func(ctx context.Context, email, passwordHash, name string) (generated.User, error)
}

func (f fakeUsersRepo) Create(ctx context.Context, email string, passwordHash string, name string) (generated.User, error) {
	return f.create(ctx, email, passwordHash, name)
}
func (f fakeUsersRepo) GetByID(ctx context.Context, id uuid.UUID) (generated.User, error) {
	return generated.User{}, pgx.ErrNoRows
}
func (f fakeUsersRepo) GetByEmail(ctx context.Context, email string) (generated.User, error) {
	return f.getByEmail(ctx, email)
}

func TestService_Register_Validation(t *testing.T) {
	svc := Service{Users: fakeUsersRepo{
		getByEmail: func(context.Context, string) (generated.User, error) { return generated.User{}, pgx.ErrNoRows },
		create:     func(context.Context, string, string, string) (generated.User, error) { return generated.User{}, nil },
	}}

	_, err := svc.Register(context.Background(), "not-an-email", "password123", "n")
	if err == nil || err != ErrInvalidEmail {
		t.Fatalf("err = %v, want %v", err, ErrInvalidEmail)
	}

	_, err = svc.Register(context.Background(), "a@b.com", "short", "n")
	if err == nil || err != ErrPasswordTooShort {
		t.Fatalf("err = %v, want %v", err, ErrPasswordTooShort)
	}
}

func TestService_Register_EmailTaken(t *testing.T) {
	svc := Service{Users: fakeUsersRepo{
		getByEmail: func(context.Context, string) (generated.User, error) { return generated.User{Email: "a@b.com"}, nil },
		create:     func(context.Context, string, string, string) (generated.User, error) { t.Fatalf("should not create"); return generated.User{}, nil },
	}}

	_, err := svc.Register(context.Background(), "a@b.com", "password123", "n")
	if err == nil || err != ErrEmailTaken {
		t.Fatalf("err = %v, want %v", err, ErrEmailTaken)
	}
}

func TestService_Register_HashesPassword(t *testing.T) {
	var capturedHash string
	svc := Service{Users: fakeUsersRepo{
		getByEmail: func(context.Context, string) (generated.User, error) { return generated.User{}, pgx.ErrNoRows },
		create: func(ctx context.Context, email, passwordHash, name string) (generated.User, error) {
			capturedHash = passwordHash
			return generated.User{ID: uuid.New(), Email: email, Name: name}, nil
		},
	}}

	pw := "password123"
	_, err := svc.Register(context.Background(), "a@b.com", pw, "n")
	if err != nil {
		t.Fatalf("Register() err = %v", err)
	}
	if capturedHash == "" || capturedHash == pw {
		t.Fatalf("capturedHash invalid: %q", capturedHash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(capturedHash), []byte(pw)); err != nil {
		t.Fatalf("hash does not match password: %v", err)
	}
}

