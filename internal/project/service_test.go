package project

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestService_CreateForWorkspace_RequiresAuth(t *testing.T) {
	svc := Service{}
	_, err := svc.CreateForWorkspace(context.Background(), uuid.Nil, uuid.New(), "p", "")
	if err != ErrUnauthenticated {
		t.Fatalf("err=%v want=%v", err, ErrUnauthenticated)
	}
}

func TestService_ListForWorkspace_RequiresAuth(t *testing.T) {
	svc := Service{}
	_, err := svc.ListForWorkspace(context.Background(), uuid.Nil, uuid.New(), 10, 0)
	if err != ErrUnauthenticated {
		t.Fatalf("err=%v want=%v", err, ErrUnauthenticated)
	}
}
