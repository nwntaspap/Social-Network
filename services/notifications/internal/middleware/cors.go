package middleware

import (
	"net/http"
	"strings"
)

type CORS struct {
	allowedOrigin string
	next          http.Handler
}

func NewCORS(allowedOrigin string, next http.Handler) *CORS {
	return &CORS{allowedOrigin: allowedOrigin, next: next}
}

func (c *CORS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" || c.allowedOrigin == "" || !c.allows(origin) {
		c.next.ServeHTTP(w, r)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	c.next.ServeHTTP(w, r)
}

func (c *CORS) allows(origin string) bool {
	if c.allowedOrigin == "*" {
		return true
	}
	for _, allowed := range strings.Split(c.allowedOrigin, ",") {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}
