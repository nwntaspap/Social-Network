package middleware

import (
	"net/http"
	"strings"
)

type corsMiddleware struct {
	handler        http.Handler
	allowedOrigins map[string]bool
}

func (c *corsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	if c.allowedOrigins[origin] {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "X-Total-Count")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	c.handler.ServeHTTP(w, r)
}

func NewCorsMiddleware(handler http.Handler) http.Handler {
	return &corsMiddleware{handler: handler}
}

func NewCorsMiddlewareWithOrigins(handler http.Handler, allowedOrigins string) http.Handler {
	origins := make(map[string]bool)
	if allowedOrigins != "" {
		for _, o := range strings.Split(allowedOrigins, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				origins[trimmed] = true
			}
		}
	}
	return &corsMiddleware{handler: handler, allowedOrigins: origins}
}
