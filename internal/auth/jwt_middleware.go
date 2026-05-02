package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ctxKey int

const userIDKey ctxKey = iota

type Middleware struct {
	JWTSecret string
}

func (m Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := m.authenticate(r)
		if !ok {
			writeError(w, toAPIError(ErrUnauthorized))
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(userIDKey)
	uid, ok := v.(uuid.UUID)
	return uid, ok
}

// ContextWithUserID attaches user_id as RequireAuth middleware would. Intended for tests and internal tooling.
func ContextWithUserID(ctx context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

func (m Middleware) authenticate(r *http.Request) (uuid.UUID, bool) {
	if m.JWTSecret == "" {
		return uuid.UUID{}, false
	}
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return uuid.UUID{}, false
	}
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return uuid.UUID{}, false
	}
	raw := strings.TrimSpace(parts[1])
	if raw == "" {
		return uuid.UUID{}, false
	}

	tok, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return []byte(m.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil || tok == nil || !tok.Valid {
		return uuid.UUID{}, false
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.UUID{}, false
	}
	userIDRaw, ok := claims["user_id"].(string)
	if !ok {
		return uuid.UUID{}, false
	}
	uid, err := uuid.Parse(userIDRaw)
	if err != nil {
		return uuid.UUID{}, false
	}
	return uid, true
}
