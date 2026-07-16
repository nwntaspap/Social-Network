package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/internal/user"
	"social-network/internal/user/commands"
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
	h := newTestHandler(func(h *Handler) {
		h.login = &stubLogin{
			result: &commands.LoginResult{
				User:  &user.User{ID: "u1", Email: "a@b.com"},
				Token: "tok123",
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

	body, _ := json.Marshal(map[string]string{"token": "tok123"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
