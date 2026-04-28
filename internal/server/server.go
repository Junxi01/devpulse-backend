package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"devpulse-backend/internal/health"
)

type Server struct {
	httpServer *http.Server
}

type Deps struct {
	Logger *slog.Logger
	DB     *pgxpool.Pool
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

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(NewSlogMiddleware(deps.Logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	h := health.Handler{DB: deps.DB}
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)

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

func NewSlogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			reqID := middleware.GetReqID(r.Context())
			attrs := []slog.Attr{
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_ip", r.RemoteAddr),
			}
			if r.UserAgent() != "" {
				attrs = append(attrs, slog.String("user_agent", r.UserAgent()))
			}

			logger.InfoContext(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path), slog.Any("http", slog.GroupValue(attrs...)))
		})
	}
}

