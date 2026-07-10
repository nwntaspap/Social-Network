package middleware

import "net/http"

type CORS struct {
	allowedOrigins map[string]bool
	wildcard       bool
	next           http.Handler
}

func NewCORS(allowedOrigins []string, next http.Handler) *CORS {
	origins := make(map[string]bool, len(allowedOrigins))
	wildcard := false
	for _, o := range allowedOrigins {
		if o == "*" {
			wildcard = true
			break
		}
		origins[o] = true
	}
	return &CORS{
		allowedOrigins: origins,
		wildcard:       wildcard,
		next:           next,
	}
}

func (c *CORS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	if origin == "" {
		c.next.ServeHTTP(w, r)
		return
	}

	allowed := c.wildcard || c.allowedOrigins[origin]
	if !allowed {
		http.Error(w, "Forbidden: origin not allowed", http.StatusForbidden)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "X-Total-Count")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	c.next.ServeHTTP(w, r)
}
