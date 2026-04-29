package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	AppMode     string
	HTTPAddr    string
	DatabaseURL string
	RedisAddr   string
	JWTSecret   string
	AIProvider  string
	GitHubMode  string
}

func Load() (Config, error) {
	// Load .env if present (no error if missing).
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:      getenv("APP_ENV", "development"),
		AppMode:     getenv("APP_MODE", "api"),
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		AIProvider:  getenv("AI_PROVIDER", "openai"),
		GitHubMode:  getenv("GITHUB_MODE", "mock"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error
	if c.HTTPAddr == "" {
		errs = append(errs, errors.New("HTTP_ADDR is required"))
	}
	if c.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if c.JWTSecret == "" {
		errs = append(errs, errors.New("JWT_SECRET is required"))
	}
	return joinErrors(errs)
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	msg := "config validation failed:"
	for _, err := range errs {
		msg += "\n- " + err.Error()
	}
	return fmt.Errorf("%s", msg)
}

