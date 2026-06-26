package middleware

import (
	"net/http"
	"time"

	"social-network/internal/domain/session"
	"social-network/internal/domain/user"
)

type Key string

const (
	userIDKey Key = "user"
)

func CheckTokenExpiration(session *session.Session) (sessionExpired, refreshTokenExpired bool) {
	if session.Expiry.Before(time.Now()) {
		sessionExpired = true
	}

	if session.RefreshTokenExpiry.Before(time.Now()) {
		refreshTokenExpired = true
	}

	return
}

func GetTokensFromRequest(r *http.Request) (sessionToken, refreshToken string) {
	cookie, err := r.Cookie("access_token")
	if err == nil {
		sessionToken = cookie.Value
	}

	cookie, err = r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	}

	// Fallback: allow token via query parameter (used by WebSocket clients
	// that cannot set Cookie headers, e.g. Postman WS, native browser WebSocket).
	if sessionToken == "" {
		if t := r.URL.Query().Get("access_token"); t != "" {
			sessionToken = t
		}
	}
	if refreshToken == "" {
		if t := r.URL.Query().Get("refresh_token"); t != "" {
			refreshToken = t
		}
	}

	return
}

func GetUserFromContext(r *http.Request) *user.User {
	value := r.Context().Value(userIDKey)
	if value == nil {
		return nil
	}

	user, ok := value.(*user.User)
	if !ok {
		return nil
	}

	return user
}
