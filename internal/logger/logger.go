package logger

import (
	"log/slog"
	"os"
)

func New(appEnv string) *slog.Logger {
	level := slog.LevelInfo
	if appEnv == "development" || appEnv == "dev" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}

