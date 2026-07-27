package middleware

import (
	"context"
	"net/http"

	"social-network/internal/core/session"
)

type Auth struct {
	sessionManager session.Manager
	cookieName     string
}

func NewAuth(sm session.Manager, cookieName string) *Auth {
	return &Auth{
		sessionManager: sm,
		cookieName:     cookieName,
	}
}

func (a *Auth) Required(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := readTokenFromRequest(r, a.cookieName)
		if token == "" {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized: missing session token")
			return
		}

		sess, err := a.sessionManager.Get(r.Context(), token)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized: invalid session")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, sess.UserID)
		ctx = context.WithValue(ctx, sessionTokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (a *Auth) Optional(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := readTokenFromRequest(r, a.cookieName)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		sess, err := a.sessionManager.Get(r.Context(), token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, sess.UserID)
		ctx = context.WithValue(ctx, sessionTokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
