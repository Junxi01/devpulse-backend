package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"devpulse-backend/internal/auth"
	"devpulse-backend/internal/health"
	appmw "devpulse-backend/internal/middleware"
)

type Server struct {
	httpServer *http.Server
}

type Deps struct {
	Logger *slog.Logger
	DB     *pgxpool.Pool
	Auth   auth.Handler
	Addr   string
}

func New(deps Deps) (*Server, error) {
	if deps.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if deps.Addr == "" {
		return nil, errors.New("addr is required")
	}

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(appmw.RequestLogger(deps.Logger))
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	h := health.Handler{DB: deps.DB}
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", deps.Auth.Register)
	})

	srv := &http.Server{
		Addr:              deps.Addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return &Server{httpServer: srv}, nil
}

func (s *Server) ListenAndServe() error {
	if s.httpServer == nil {
		return errors.New("server not initialized")
	}
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}
	return s.httpServer.Serve(ln)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

