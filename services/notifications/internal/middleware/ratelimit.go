package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"social-network/services/notifications/internal/middleware/ratelimiter"
)

type RateLimiter struct {
	window *ratelimiter.Window
	limit  int
	next   http.Handler
}

func NewRateLimiter(limit int, window time.Duration, next http.Handler) *RateLimiter {
	return &RateLimiter{
		window: ratelimiter.NewWindow(limit, window),
		limit:  limit,
		next:   next,
	}
}

func (rl *RateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := getClientIP(r)

	allowed, remaining, resetTime := rl.window.Allow(ip)

	w.Header().Set("X-Ratelimit-Limit", strconv.Itoa(rl.limit))
	w.Header().Set("X-Ratelimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-Ratelimit-Reset", strconv.FormatInt(resetTime, 10))

	if !allowed {
		retryAfter := max(resetTime-time.Now().Unix(), 1)
		w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
		RespondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded, try again later")
		return
	}

	rl.next.ServeHTTP(w, r)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
