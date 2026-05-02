package workspace

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestService_Create_RequiresAuth(t *testing.T) {
	svc := Service{}
	_, err := svc.Create(context.Background(), uuid.Nil, "my ws")
	if err != ErrUnauthenticated {
		t.Fatalf("err=%v want=%v", err, ErrUnauthenticated)
	}
}

func TestService_Create_RequiresName(t *testing.T) {
	svc := Service{}
	_, err := svc.Create(context.Background(), uuid.New(), "   ")
	if err != ErrInvalidName {
		t.Fatalf("err=%v want=%v", err, ErrInvalidName)
	}
}

func TestService_List_RequiresAuth(t *testing.T) {
	svc := Service{}
	_, err := svc.ListByUser(context.Background(), uuid.Nil, 10, 0)
	if err != ErrUnauthenticated {
		t.Fatalf("err=%v want=%v", err, ErrUnauthenticated)
	}
}

func TestService_Get_RequiresAuth(t *testing.T) {
	svc := Service{}
	_, err := svc.GetByID(context.Background(), uuid.Nil, uuid.New())
	if err != ErrUnauthenticated {
		t.Fatalf("err=%v want=%v", err, ErrUnauthenticated)
	}
}

