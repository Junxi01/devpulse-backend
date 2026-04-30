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
	"devpulse-backend/internal/auth"
	"devpulse-backend/internal/db"
	"devpulse-backend/internal/logger"
	"devpulse-backend/internal/repository"
	"devpulse-backend/internal/server"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	appLogger := logger.New(cfg.AppEnv)
	slog.SetDefault(appLogger)

	var dbPool *pgxpool.Pool
	var usersRepo repository.UserRepository
	database, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		// Allow the API to start even when DB is unavailable; /readyz will reflect readiness.
		appLogger.Warn("database unavailable; readiness will be false until it recovers", slog.String("error", err.Error()))
	} else {
		dbPool = database.Pool
		usersRepo = repository.NewUserRepository(database.Queries)
		defer database.Pool.Close()
	}

	authSvc := auth.Service{Users: usersRepo}
	authHandler := auth.Handler{Service: authSvc}

	srv, err := server.New(server.Deps{
		Logger: appLogger,
		DB:     dbPool,
		Auth:   authHandler,
		Addr:   cfg.HTTPAddr,
	})
	if err != nil {
		appLogger.Error("failed to initialize http server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		appLogger.Info("http server starting", slog.String("addr", cfg.HTTPAddr), slog.String("env", cfg.AppEnv), slog.String("mode", cfg.AppMode))
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		appLogger.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
			appLogger.Error("http server stopped", slog.String("error", err.Error()))
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("http server shutdown error", slog.String("error", err.Error()))
	}
}

