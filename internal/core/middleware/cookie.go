package middleware

import (
	"net/http"
	"time"
)

type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

type SessionCookies struct {
	cfg CookieConfig
}

func NewSessionCookies(cfg CookieConfig) *SessionCookies {
	return &SessionCookies{cfg: cfg}
}

func (c *SessionCookies) SetAccessCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{ // #nosec G124 -- fields set from config at runtime
		Name:     c.cfg.Name,
		Value:    token,
		Path:     c.cfg.Path,
		Domain:   c.cfg.Domain,
		HttpOnly: c.cfg.HTTPOnly,
		Secure:   c.cfg.Secure,
		SameSite: c.cfg.SameSite,
		Expires:  expiresAt.UTC(),
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	})
}

func (c *SessionCookies) DeleteAccessCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{ // #nosec G124 -- clearing cookie, not setting auth
		Name:     c.cfg.Name,
		Value:    "",
		Path:     c.cfg.Path,
		Domain:   c.cfg.Domain,
		HttpOnly: true,
		Secure:   c.cfg.Secure,
		SameSite: c.cfg.SameSite,
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
	})
}

func ParseSameSite(s string) http.SameSite {
	switch s {
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
