package authcookies

import (
	"net/http"
	"time"

	"social-network/internal/config"
	"social-network/internal/domain/session"
)

type Manager struct {
	cfg config.SessionManagerConfig
}

func NewManager(cfg config.SessionManagerConfig) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) ReadTokens(r *http.Request) (sessionToken, refreshToken string) {
	cookie, err := r.Cookie(m.cfg.AccessCookieName)
	if err == nil {
		sessionToken = cookie.Value
	}

	cookie, err = r.Cookie(m.cfg.RefreshCookieName)
	if err == nil {
		refreshToken = cookie.Value
	}

	// Fallback: allow token via query parameter (used by WebSocket clients
	// that cannot set Cookie headers, e.g. Postman WS, native browser WebSocket).
	if sessionToken == "" {
		if t := r.URL.Query().Get(m.cfg.AccessCookieName); t != "" {
			sessionToken = t
		}
	}
	if refreshToken == "" {
		if t := r.URL.Query().Get(m.cfg.RefreshCookieName); t != "" {
			refreshToken = t
		}
	}

	return
}

func (m *Manager) SetCookies(w http.ResponseWriter, session *session.Session) {
	accessCookie, refreshCookie := m.NewSessionCookie(session)

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
}

func (m *Manager) DeleteCookies(r *http.Request, w http.ResponseWriter) (sessiontoken string) {
	cookie, err := r.Cookie(m.cfg.AccessCookieName)
	if err == nil {
		sessiontoken = cookie.Value
		http.SetCookie(w, &http.Cookie{
			Name:     m.cfg.AccessCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})
	}

	cookie, err = r.Cookie(m.cfg.RefreshCookieName)
	if err == nil {
		cookie.MaxAge = -1
		http.SetCookie(w, &http.Cookie{
			Name:     m.cfg.RefreshCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})
	}
	return sessiontoken
}

func parseSameSite(s string) http.SameSite {
	switch s {
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func (m *Manager) NewSessionCookie(session *session.Session) (accessCookie, refreshCookie *http.Cookie) {
	accessMaxAge := int(time.Until(session.Expiry).Seconds())
	if accessMaxAge < 0 {
		accessMaxAge = 0
	}

	refreshMaxAge := int(time.Until(session.RefreshTokenExpiry).Seconds())
	if refreshMaxAge < 0 {
		refreshMaxAge = 0
	}

	return &http.Cookie{
			Name:     m.cfg.AccessCookieName,
			Value:    session.AccessToken,
			Path:     m.cfg.CookiePath,
			Domain:   m.cfg.CookieDomain,
			HttpOnly: m.cfg.HTTPOnlyCookie,
			Secure:   m.cfg.SecureCookie,
			SameSite: parseSameSite(m.cfg.SameSite),
			Expires:  session.Expiry.UTC(),
			MaxAge:   accessMaxAge,
		},
		&http.Cookie{
			Name:     m.cfg.RefreshCookieName,
			Value:    session.RefreshToken,
			Path:     m.cfg.CookiePath,
			Domain:   m.cfg.CookieDomain,
			HttpOnly: m.cfg.HTTPOnlyCookie,
			Secure:   m.cfg.SecureCookie,
			SameSite: parseSameSite(m.cfg.SameSite),
			Expires:  session.RefreshTokenExpiry.UTC(),
			MaxAge:   refreshMaxAge,
		}
}
