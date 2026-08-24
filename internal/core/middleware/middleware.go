package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

type contextIDKey string

const (
	userIDKey       contextIDKey = "UserID"
	sessionTokenKey contextIDKey = "SessionToken"
)

func GetUserIDFromContext(r *http.Request) string {
	v, _ := r.Context().Value(userIDKey).(string)
	return v
}

func GetSessionTokenFromContext(r *http.Request) string {
	v, _ := r.Context().Value(sessionTokenKey).(string)
	return v
}

// WithSessionToken injects a session token into the request context.
// Intended for use in tests where the auth middleware is bypassed.
func WithSessionToken(r *http.Request, token string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), sessionTokenKey, token))
}

func readTokenFromRequest(r *http.Request, cookieName string) string {
	if c, err := r.Cookie(cookieName); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
