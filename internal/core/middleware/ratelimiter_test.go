package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/core/middleware/ratelimiter"
	"social-network/internal/platform/cache"
)

func TestRateLimiter_Allow(t *testing.T) {
	c := cache.NewInMemoryCache()
	defer c.Stop()

	w := ratelimiter.NewWindow(c, 10, 60*time.Second)
	rl := NewRateLimiter(w, 10, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	w2 := httptest.NewRecorder()

	rl.ServeHTTP(w2, r)

	if w2.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w2.Code, http.StatusOK)
	}
	if got := w2.Header().Get("X-Ratelimit-Limit"); got != "10" {
		t.Errorf("X-Ratelimit-Limit = %q, want %q", got, "10")
	}
	if got := w2.Header().Get("X-Ratelimit-Remaining"); got != "9" {
		t.Errorf("X-Ratelimit-Remaining = %q, want %q", got, "9")
	}
}

func TestRateLimiter_Block(t *testing.T) {
	c := cache.NewInMemoryCache()
	defer c.Stop()

	w := ratelimiter.NewWindow(c, 2, 60*time.Second)
	rl := NewRateLimiter(w, 2, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the limit
	for range 2 {
		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.1:1234"
		rl.ServeHTTP(httptest.NewRecorder(), r)
	}

	// This one should be blocked
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	w2 := httptest.NewRecorder()

	rl.ServeHTTP(w2, r)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w2.Code, http.StatusTooManyRequests)
	}
	if got := w2.Header().Get("Retry-After"); got == "" {
		t.Error("Retry-After header missing")
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")

	got := getClientIP(r)
	if got != "1.2.3.4" {
		t.Errorf("getClientIP() = %q, want %q", got, "1.2.3.4")
	}
}

func TestGetClientIP_XRealIP(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.Header.Set("X-Real-IP", "9.8.7.6")

	got := getClientIP(r)
	if got != "9.8.7.6" {
		t.Errorf("getClientIP() = %q, want %q", got, "9.8.7.6")
	}
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.RemoteAddr = "4.3.2.1:54321"

	got := getClientIP(r)
	if got != "4.3.2.1" {
		t.Errorf("getClientIP() = %q, want %q", got, "4.3.2.1")
	}
}
