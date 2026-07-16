package middleware

import (
	"encoding/json"
	"net/http"
)

type contextIDKey string

const userIDKey contextIDKey = "UserID"

func GetUserIDFromContext(r *http.Request) string {
	v, _ := r.Context().Value(userIDKey).(string)
	return v
}

func readTokenFromRequest(r *http.Request, cookieName string) string {
	if c, err := r.Cookie(cookieName); err == nil && c.Value != "" {
		return c.Value
	}
	if t := r.URL.Query().Get(cookieName); t != "" {
		return t
	}
	return ""
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
