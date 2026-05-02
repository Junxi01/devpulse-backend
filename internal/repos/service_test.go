package repos

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestService_CreateForProject_RequiresAuth(t *testing.T) {
	svc := Service{}
	_, err := svc.CreateForProject(context.Background(), uuid.Nil, uuid.New(), "github", "o", "n", "o/n", "1", "main")
	if err != ErrUnauthenticated {
		t.Fatalf("err=%v want=%v", err, ErrUnauthenticated)
	}
}

func TestService_ListForProject_RequiresAuth(t *testing.T) {
	svc := Service{}
	_, err := svc.ListForProject(context.Background(), uuid.Nil, uuid.New(), 10, 0)
	if err != ErrUnauthenticated {
		t.Fatalf("err=%v want=%v", err, ErrUnauthenticated)
	}
}
