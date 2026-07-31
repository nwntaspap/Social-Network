package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionCookies_SetAccessCookie(t *testing.T) {
	c := NewSessionCookies(CookieConfig{
		Name:     "access_token",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)

	rr := httptest.NewRecorder()
	c.SetAccessCookie(rr, "tok123", expiresAt)

	got := findCookie(t, rr, "access_token")
	if got.Value != "tok123" {
		t.Errorf("cookie value = %q, want tok123", got.Value)
	}
	if got.MaxAge == 0 {
		t.Error("cookie MaxAge not set")
	}
	if !got.Expires.Equal(expiresAt) {
		t.Errorf("cookie expires = %v, want %v", got.Expires, expiresAt)
	}
}

func TestSessionCookies_DeleteAccessCookie(t *testing.T) {
	c := NewSessionCookies(CookieConfig{Name: "access_token", Path: "/"})

	rr := httptest.NewRecorder()
	c.DeleteAccessCookie(rr)

	got := findCookie(t, rr, "access_token")
	if got.Value != "" {
		t.Errorf("cookie value = %q, want empty", got.Value)
	}
	if got.MaxAge != -1 {
		t.Errorf("cookie MaxAge = %d, want -1", got.MaxAge)
	}
}

func findCookie(t *testing.T, rr *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, c := range rr.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("cookie %q not set", name)
	return nil
}
