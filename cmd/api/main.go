package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devpulse-backend/internal/config"
	"devpulse-backend/internal/db"
	"devpulse-backend/internal/server"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	logger := newLogger(cfg.AppEnv)
	slog.SetDefault(logger)

	database, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Pool.Close()

	srv, err := server.New(server.Deps{
		Logger: logger,
		DB:     database.Pool,
		Addr:   cfg.HTTPAddr,
	})
	if err != nil {
		logger.Error("failed to initialize http server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server starting", slog.String("addr", cfg.HTTPAddr), slog.String("env", cfg.AppEnv))
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
			logger.Error("http server stopped", slog.String("error", err.Error()))
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown error", slog.String("error", err.Error()))
	}
}

func newLogger(appEnv string) *slog.Logger {
	level := slog.LevelInfo
	if appEnv == "development" || appEnv == "dev" {
		level = slog.LevelDebug
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(h)
}

