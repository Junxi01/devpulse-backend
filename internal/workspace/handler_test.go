package workspace

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Create_UnauthorizedWithoutContextUser(t *testing.T) {
	h := Handler{}
	req := httptest.NewRequest(http.MethodPost, "/workspaces", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Create(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

