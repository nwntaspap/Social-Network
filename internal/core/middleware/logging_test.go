package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogging_RecordsStatusCode(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	logging := NewLogging(logger, inner)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/missing", nil)
	w := httptest.NewRecorder()

	logging.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, `"status":404`) {
		t.Errorf("log missing status 404: %s", logOutput)
	}
}

func TestLogging_GeneratesRequestID(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	logging := NewLogging(logger, inner)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	logging.ServeHTTP(w, r)

	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Error("X-Request-ID header not set")
	}
}

func TestLogging_PreservesExistingRequestID(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	logging := NewLogging(logger, inner)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.Header.Set("X-Request-ID", "my-custom-id")
	w := httptest.NewRecorder()

	logging.ServeHTTP(w, r)

	rid := w.Header().Get("X-Request-ID")
	if rid != "my-custom-id" {
		t.Errorf("X-Request-ID = %q, want %q", rid, "my-custom-id")
	}
}
