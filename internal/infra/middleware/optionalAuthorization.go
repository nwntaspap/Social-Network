package middleware

import (
	"context"
	"net/http"

	"social-network/internal/domain/session"
)

// type optionalAuthMiddleware struct {
// 	sessionManager session.Manager
// }

// type OptionalAuthMiddleware interface {
// 	Optional(next http.HandlerFunc) http.HandlerFunc
// }

// func NewOptionalAuthMiddleware(sessionManager session.Manager) AuthorizationInterface {
// 	return Authorization{
// 		sessionManager: sessionManager,
// 	}
// }

func (a authorization) Optional(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionToken, refreshToken := a.cookieManager.ReadTokens(r)
		if sessionToken == "" && refreshToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		var (
			session *session.Session
			err     error
		)

		if sessionToken == "" && refreshToken != "" {
			session, err = a.sessionManager.GetSessionByRefreshToken(refreshToken)
		} else {
			session, err = a.sessionManager.GetSessionFromSessionTokens(sessionToken, refreshToken)
		}
		if err != nil || session == nil {
			next.ServeHTTP(w, r)
			return
		}

		sessionExpired, refreshTokenExpired := CheckTokenExpiration(session)

		// If refresh token expired, continue without authentication
		if refreshTokenExpired {
			next.ServeHTTP(w, r)
			return
		}

		// If access token expired but refresh is valid, create new session
		if sessionExpired {
			_ = a.sessionManager.DeleteSession(session.AccessToken)
			session, err = a.sessionManager.CreateSession(r.Context(), session.UserID)
			if err != nil || session == nil {
				next.ServeHTTP(w, r)
				return
			}
			a.cookieManager.SetCookies(w, session)
		}

		// Both tokens valid, or just refreshed - get user and proceed with context
		user, err := a.sessionManager.GetUserFromSession(session.AccessToken)
		if err != nil || user == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
