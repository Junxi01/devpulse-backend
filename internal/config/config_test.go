package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_MODE", "")
	t.Setenv("HTTP_ADDR", "")

	// Required fields for validation.
	t.Setenv("DATABASE_URL", "postgres://devpulse:devpulse@postgres:5432/devpulse?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8080")
	}
	if cfg.AppMode != "demo" {
		t.Fatalf("AppMode = %q, want %q", cfg.AppMode, "demo")
	}
	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "development")
	}

	// Ensure Load does not clobber required values.
	if cfg.DatabaseURL != os.Getenv("DATABASE_URL") {
		t.Fatalf("DatabaseURL = %q, want %q", cfg.DatabaseURL, os.Getenv("DATABASE_URL"))
	}
}

func TestLoad_AppModeDemoCaseInsensitive(t *testing.T) {
	t.Setenv("APP_MODE", "DEMO")
	t.Setenv("DATABASE_URL", "postgres://x:x@localhost:5432/x?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("HTTP_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppMode != "demo" {
		t.Fatalf("AppMode = %q, want demo", cfg.AppMode)
	}
}

func TestLoad_InvalidAppMode(t *testing.T) {
	t.Setenv("APP_MODE", "staging")
	t.Setenv("DATABASE_URL", "postgres://x:x@localhost:5432/x?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid APP_MODE")
	}
	if !strings.Contains(err.Error(), "APP_MODE") {
		t.Fatalf("unexpected error: %v", err)
	}
}
