package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "UserID"

func UserIDKey() contextKey {
	return userIDKey
}

func GetUserID(r *http.Request) string {
	v, _ := r.Context().Value(userIDKey).(string)
	return v
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type StatusRecorder struct {
	http.ResponseWriter
	status int
}

type Logging struct {
	logger *slog.Logger
	next   http.Handler
}

func NewLogging(logger *slog.Logger, next http.Handler) *Logging {
	return &Logging{logger: logger, next: next}
}

func (l *Logging) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.NewString()
	}

	rec := &StatusRecorder{ResponseWriter: w, status: http.StatusOK}
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
