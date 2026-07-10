package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type StatusRecorder struct {
	http.ResponseWriter

	status int
}

type Logging struct {
	logger *slog.Logger
	next   http.Handler
}

func NewLogging(logger *slog.Logger, next http.Handler) *Logging {
	return &Logging{
		logger: logger,
		next:   next,
	}
}

func (l *Logging) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.NewString()
	}

	rec := &StatusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}

	l.next.ServeHTTP(rec, r)

	l.logger.Info(
		"request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", rec.status),
		slog.Duration("duration", time.Since(start)),
		slog.String("request_id", requestID),
	)

	w.Header().Set("X-Request-ID", requestID)
}

func (r *StatusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
