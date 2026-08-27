package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/core/session"
)

var errNotImplemented = errors.New("not implemented")

func TestGetUserIDFromContext_Present(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), userIDKey, "user-42"))

	got := GetUserIDFromContext(r)
	if got != "user-42" {
		t.Errorf("GetUserIDFromContext() = %q, want %q", got, "user-42")
	}
}

func TestGetUserIDFromContext_Missing(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	got := GetUserIDFromContext(r)
	if got != "" {
		t.Errorf("GetUserIDFromContext() = %q, want empty string", got)
	}
}

func TestReadTokenFromRequest_Cookie(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: "tok-123", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})

	got := readTokenFromRequest(r, "session_token")
	if got != "tok-123" {
		t.Errorf("readTokenFromRequest() = %q, want %q", got, "tok-123")
	}
}

func TestReadTokenFromRequest_QueryParamRejected(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?session_token=tok-456", nil)

	got := readTokenFromRequest(r, "session_token")
	if got != "" {
		t.Errorf("readTokenFromRequest() = %q, want empty string (query params must not authenticate)", got)
	}
}

func TestReadTokenFromRequest_Empty(t *testing.T) {
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	got := readTokenFromRequest(r, "session_token")
	if got != "" {
		t.Errorf("readTokenFromRequest() = %q, want empty string", got)
	}
}

func TestRespondWithError(t *testing.T) {
	w := httptest.NewRecorder()

	respondWithError(w, http.StatusUnauthorized, "Invalid session")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != "Invalid session" {
		t.Errorf("body error = %q, want %q", body["error"], "Invalid session")
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

// --- Mock session manager ---

type mockSessionManager struct {
	getFn func(ctx context.Context, token string) (*session.Session, error)
}

func (m *mockSessionManager) Create(_ context.Context, _ string) (*session.Session, error) {
	return nil, errNotImplemented
}

func (m *mockSessionManager) Get(ctx context.Context, token string) (*session.Session, error) {
	if m.getFn != nil {
		return m.getFn(ctx, token)
	}
	return nil, errors.New("not found")
}

func (m *mockSessionManager) Revoke(_ context.Context, _ string) error { return errNotImplemented }

// --- Auth Required tests ---

func TestAuth_Required_ValidSession(t *testing.T) {
	sm := &mockSessionManager{
		getFn: func(_ context.Context, token string) (*session.Session, error) {
			if token == "valid-token" {
				return &session.Session{
					Token:     "valid-token",
					UserID:    "user-42",
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil
			}
			return nil, errors.New("not found")
		},
	}

	auth := NewAuth(sm, "session_token")

	handler := auth.Required(func(w http.ResponseWriter, r *http.Request) {
		uid := GetUserIDFromContext(r)
		_, _ = w.Write([]byte(uid))
	})

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	w := httptest.NewRecorder()

	handler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "user-42" {
		t.Errorf("body = %q, want %q", w.Body.String(), "user-42")
	}
}

func TestAuth_Required_NoToken(t *testing.T) {
	sm := &mockSessionManager{}
	auth := NewAuth(sm, "session_token")

	handler := auth.Required(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_Required_InvalidToken(t *testing.T) {
	sm := &mockSessionManager{
		getFn: func(_ context.Context, _ string) (*session.Session, error) {
			return nil, errors.New("session expired")
		},
	}
	auth := NewAuth(sm, "session_token")

	handler := auth.Required(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: "bad-token", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	w := httptest.NewRecorder()

	handler(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_Required_QueryParamRejected(t *testing.T) {
	sm := &mockSessionManager{
		getFn: func(_ context.Context, token string) (*session.Session, error) {
			if token == "ws-token" {
				return &session.Session{
					Token:     "ws-token",
					UserID:    "user-99",
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil
			}
			return nil, errors.New("not found")
		},
	}
	auth := NewAuth(sm, "access_token")

	handler := auth.Required(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?access_token=ws-token", nil)
	w := httptest.NewRecorder()

	handler(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// --- Auth Optional tests ---

func TestAuth_Optional_ValidSession(t *testing.T) {
	sm := &mockSessionManager{
		getFn: func(_ context.Context, token string) (*session.Session, error) {
			if token == "valid-token" {
				return &session.Session{
					Token:     "valid-token",
					UserID:    "user-77",
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil
			}
			return nil, errors.New("not found")
		},
	}
	auth := NewAuth(sm, "session_token")

	handler := auth.Optional(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(GetUserIDFromContext(r)))
	})

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	w := httptest.NewRecorder()

	handler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "user-77" {
		t.Errorf("body = %q, want %q", w.Body.String(), "user-77")
	}
}

func TestAuth_Optional_NoToken(t *testing.T) {
	sm := &mockSessionManager{}
	auth := NewAuth(sm, "session_token")

	called := false
	handler := auth.Optional(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if uid := GetUserIDFromContext(r); uid != "" {
			t.Errorf("userID in context = %q, want empty", uid)
		}
	})

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, r)

	if !called {
		t.Fatal("downstream handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuth_Optional_InvalidToken(t *testing.T) {
	sm := &mockSessionManager{
		getFn: func(_ context.Context, _ string) (*session.Session, error) {
			return nil, errors.New("not found")
		},
	}
	auth := NewAuth(sm, "session_token")

	called := false
	handler := auth.Optional(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "session_token", Value: "bad-token", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	w := httptest.NewRecorder()

	handler(w, r)

	if !called {
		t.Fatal("downstream handler was not called for optional auth with invalid token")
	}
}
