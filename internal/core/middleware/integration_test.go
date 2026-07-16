package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/core/middleware/ratelimiter"
	"social-network/internal/core/session"
	"social-network/internal/platform/cache"
)

func TestMiddlewareChain_Logging_CORS_Auth(t *testing.T) {
	sm := &mockSessionManager{
		getFn: func(_ context.Context, token string) (*session.Session, error) {
			if token == "valid-token" {
				return &session.Session{
					Token:     "valid-token",
					UserID:    "user-42",
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil
			}
			return nil, errors.New("not found")
		},
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := GetUserIDFromContext(r)
		_, _ = w.Write([]byte(uid))
	})

	logger := slog.New(slog.DiscardHandler)
	auth := NewAuth(sm, "access_token")
	cors := NewCORS([]string{"https://example.com"}, auth.Required(final))
	logging := NewLogging(logger, cors)

	// Valid origin + valid token → 200, userID in body
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users/me", nil)
	r.Header.Set("Origin", "https://example.com")
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "valid-token", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	w := httptest.NewRecorder()

	logging.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("valid request: status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "user-42" {
		t.Errorf("valid request: body = %q, want %q", w.Body.String(), "user-42")
	}
	if rid := w.Header().Get("X-Request-ID"); rid == "" {
		t.Error("valid request: X-Request-ID not set")
	}

	// Disallowed origin → 403
	r = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users/me", nil)
	r.Header.Set("Origin", "https://evil.com")
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "valid-token", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	w = httptest.NewRecorder()

	logging.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("bad origin: status = %d, want %d", w.Code, http.StatusForbidden)
	}

	// Valid origin, no token → 401
	r = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users/me", nil)
	r.Header.Set("Origin", "https://example.com")
	w = httptest.NewRecorder()

	logging.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("no token: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMiddlewareChain_RateLimiter(t *testing.T) {
	c := cache.NewInMemoryCache()
	defer c.Stop()

	w := ratelimiter.NewWindow(c, 2, 60*time.Second)

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rl := NewRateLimiter(w, 2, final)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"

	// Request 1: allowed
	rec := httptest.NewRecorder()
	rl.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Errorf("request 1: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Request 2: allowed
	rec = httptest.NewRecorder()
	rl.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Errorf("request 2: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Request 3: blocked
	rec = httptest.NewRecorder()
	rl.ServeHTTP(rec, r)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("request 3: status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if got := rec.Header().Get("Retry-After"); got == "" {
		t.Error("rate limited: Retry-After header missing")
	}
}
