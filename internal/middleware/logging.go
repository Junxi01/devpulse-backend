package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			reqID := chimw.GetReqID(r.Context())
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

