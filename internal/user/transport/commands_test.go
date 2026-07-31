package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/core/middleware"
	"social-network/internal/user"
	"social-network/internal/user/commands"
	"social-network/internal/user/queries"
)

func TestRegister_Success(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.register = &stubRegister{
			user: &user.User{ID: "u1", Email: "a@b.com"},
		}
	})
	withDefaults(h)

	body, _ := json.Marshal(map[string]string{
		"email":       "a@b.com",
		"password":    "password123",
		"nickname":    "nick",
		"dateOfBirth": "2000-01-01T00:00:00Z",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	data, _ := json.Marshal(resp["data"])
	var dataMap map[string]any
	_ = json.Unmarshal(data, &dataMap)
	if dataMap["id"] != "u1" {
		t.Errorf("data.id = %v, want u1", dataMap["id"])
	}
}

func TestRegister_DateOnlyAccepted(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.register = &stubRegister{user: &user.User{ID: "u1", Email: "a@b.com"}}
	})
	withDefaults(h)

	body, _ := json.Marshal(map[string]string{
		"email":       "a@b.com",
		"password":    "password123",
		"nickname":    "nick",
		"dateOfBirth": "2000-01-15",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	stub, ok := h.register.(*stubRegister)
	if !ok {
		t.Fatal("register handler is not a *stubRegister")
	}
	got := stub.got
	want := time.Date(2000, 1, 15, 0, 0, 0, 0, time.UTC)
	if !got.DateOfBirth.Equal(want) {
		t.Errorf("dateOfBirth = %v, want %v", got.DateOfBirth, want)
	}
}

func TestRegister_InvalidDateFormat(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.register = &stubRegister{user: &user.User{ID: "u1", Email: "a@b.com"}}
	})
	withDefaults(h)

	body, _ := json.Marshal(map[string]string{
		"email":       "a@b.com",
		"password":    "password123",
		"nickname":    "nick",
		"dateOfBirth": "not-a-date",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestRegister_MethodNotAllowed(t *testing.T) {
	h := newTestHandler()
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/register", nil)
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestRegister_InvalidBody(t *testing.T) {
	h := newTestHandler()
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/register", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.register = &stubRegister{err: commands.ErrEmailTaken}
	})
	withDefaults(h)

	body, _ := json.Marshal(map[string]string{
		"email":       "a@b.com",
		"password":    "password123",
		"nickname":    "nick",
		"dateOfBirth": "2000-01-01T00:00:00Z",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusConflict)
	}
}

func TestRegister_InternalError(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.register = &stubRegister{err: errors.New("db down")}
	})
	withDefaults(h)

	body, _ := json.Marshal(map[string]string{
		"email":       "a@b.com",
		"password":    "password123",
		"nickname":    "nick",
		"dateOfBirth": "2000-01-01T00:00:00Z",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestLogin_Success(t *testing.T) {
	createdAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	expiresAt := createdAt.Add(24 * time.Hour)
	h := newTestHandler(func(h *Handler) {
		h.login = &stubLogin{
			result: &commands.LoginResult{
				User: &user.User{
					ID:          "u1",
					Email:       "a@b.com",
					FirstName:   "Alice",
					LastName:    "Smith",
					DateOfBirth: time.Date(1990, 5, 20, 0, 0, 0, 0, time.UTC),
					Nickname:    "alice",
					AboutMe:     "hello there",
					AvatarPath:  "/avatars/a.png",
					IsPrivate:   true,
					CreatedAt:   createdAt,
				},
				Token:     "tok123",
				ExpiresAt: expiresAt,
			},
		}
	})
	withDefaults(h)

	body, _ := json.Marshal(map[string]string{
		"identifier": "a@b.com",
		"password":   "pass",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	data, _ := json.Marshal(resp["data"])
	var dataMap map[string]any
	_ = json.Unmarshal(data, &dataMap)
	if dataMap["token"] != "tok123" {
		t.Errorf("data.token = %v, want tok123", dataMap["token"])
	}

	userMap, ok := dataMap["user"].(map[string]any)
	if !ok {
		t.Fatalf("data.user = %#v, want map[string]any", dataMap["user"])
	}
	wantUser := map[string]any{
		"id":          "u1",
		"email":       "a@b.com",
		"username":    "alice",
		"nickname":    "alice",
		"firstName":   "Alice",
		"lastName":    "Smith",
		"aboutMe":     "hello there",
		"dateOfBirth": "1990-05-20",
		"avatarUrl":   "/avatars/a.png",
		"isPublic":    false,
		"createdAt":   createdAt.Format(time.RFC3339),
	}
	for k, want := range wantUser {
		if userMap[k] != want {
			t.Errorf("data.user.%s = %v, want %v", k, userMap[k], want)
		}
	}
}

func TestLogin_SetsAccessCookie(t *testing.T) {
	expiresAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC).Add(24 * time.Hour)
	writer := &stubCookieWriter{}
	h := newTestHandler(func(h *Handler) {
		h.login = &stubLogin{
			result: &commands.LoginResult{
				User:      &user.User{ID: "u1", Email: "a@b.com"},
				Token:     "tok123",
				ExpiresAt: expiresAt,
			},
		}
		h.sessionCookies = writer
	})
	withDefaults(h)

	body, _ := json.Marshal(map[string]string{
		"identifier": "a@b.com",
		"password":   "pass",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !writer.setCalled {
		t.Fatal("SetAccessCookie not called")
	}
	if writer.setToken != "tok123" {
		t.Errorf("SetAccessCookie token = %q, want tok123", writer.setToken)
	}
	if !writer.setExpiry.Equal(expiresAt) {
		t.Errorf("SetAccessCookie expiry = %v, want %v", writer.setExpiry, expiresAt)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.login = &stubLogin{err: commands.ErrInvalidCredentials}
	})
	withDefaults(h)

	body, _ := json.Marshal(map[string]string{
		"identifier": "a@b.com",
		"password":   "wrong",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestLogin_InternalError(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.login = &stubLogin{err: errors.New("db down")}
	})
	withDefaults(h)

	body, _ := json.Marshal(map[string]string{
		"identifier": "a@b.com",
		"password":   "pass",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestLogout_Success(t *testing.T) {
	h := newTestHandler()
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/logout", nil)
	req = middleware.WithSessionToken(req, "tok123")
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestLogout_DeletesCookie(t *testing.T) {
	writer := &stubCookieWriter{}
	h := newTestHandler(func(h *Handler) {
		h.sessionCookies = writer
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/logout", nil)
	req = middleware.WithSessionToken(req, "tok123")
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !writer.deleteCalled {
		t.Error("DeleteAccessCookie not called")
	}
}

func TestGetMe_Success(t *testing.T) {
	createdAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	h := newTestHandler(func(h *Handler) {
		h.getProfile = &stubGetProfile{
			result: &queries.ProfileResult{
				User: user.User{
					ID:          "u1",
					Email:       "a@b.com",
					FirstName:   "Alice",
					LastName:    "Smith",
					DateOfBirth: time.Date(1990, 5, 20, 0, 0, 0, 0, time.UTC),
					Nickname:    "alice",
					AboutMe:     "hello there",
					AvatarPath:  "/avatars/a.png",
					IsPrivate:   true,
					CreatedAt:   createdAt,
				},
			},
		}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/me", nil)
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	data, _ := json.Marshal(resp["data"])
	var userMap map[string]any
	_ = json.Unmarshal(data, &userMap)

	wantUser := map[string]any{
		"id":          "u1",
		"email":       "a@b.com",
		"username":    "alice",
		"nickname":    "alice",
		"firstName":   "Alice",
		"lastName":    "Smith",
		"aboutMe":     "hello there",
		"dateOfBirth": "1990-05-20",
		"avatarUrl":   "/avatars/a.png",
		"isPublic":    false,
		"createdAt":   createdAt.Format(time.RFC3339),
	}
	for k, want := range wantUser {
		if userMap[k] != want {
			t.Errorf("data.%s = %v, want %v", k, userMap[k], want)
		}
	}
}

func TestGetMe_Unauthorized(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.auth = &stubAuth{ok: false}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/me", nil)
	rr := httptest.NewRecorder()

	h.GetMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}
