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
)

func TestUpdateProfile_Success(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.updateProfile = &stubUpdateProfile{}
	})
	withDefaults(h)
	h.auth = &stubAuth{userID: "u1", ok: true}

	body, _ := json.Marshal(map[string]string{
		"firstName":   "Jane",
		"lastName":    "Doe",
		"nickname":    "jane",
		"aboutMe":     "hello",
		"dateOfBirth": "2000-05-17",
		"gender":      "female",
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	stub, ok := h.updateProfile.(*stubUpdateProfile)
	if !ok || stub.got == nil {
		t.Fatal("UpdateProfile was never executed with a stubUpdateProfile")
	}
	got := stub.got
	if got.DateOfBirth.Format(dateOnlyLayout) != "2000-05-17" {
		t.Errorf("DateOfBirth = %v, want 2000-05-17", got.DateOfBirth)
	}
	if got.Gender != "female" {
		t.Errorf("Gender = %q, want %q", got.Gender, "female")
	}
}

func TestUpdateProfile_InvalidDOB(t *testing.T) {
	h := newTestHandler()
	withDefaults(h)
	h.auth = &stubAuth{userID: "u1", ok: true}

	body, _ := json.Marshal(map[string]string{"dateOfBirth": "not-a-date"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateProfile_Unauthorized(t *testing.T) {
	h := newTestHandler()
	withDefaults(h)
	h.auth = &stubAuth{ok: false}

	body, _ := json.Marshal(map[string]string{"firstName": "Jane"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestUpdateProfile_InternalError(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.updateProfile = &stubUpdateProfile{err: errors.New("db down")}
	})
	withDefaults(h)
	h.auth = &stubAuth{userID: "u1", ok: true}

	body, _ := json.Marshal(map[string]string{"firstName": "Jane"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestUpdateProfile_NotFound(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.updateProfile = &stubUpdateProfile{err: user.ErrUserNotFound}
	})
	withDefaults(h)
	h.auth = &stubAuth{userID: "u1", ok: true}

	body, _ := json.Marshal(map[string]string{"firstName": "Jane"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestTogglePrivacy_Success(t *testing.T) {
	h := newTestHandler()
	withDefaults(h)
	h.auth = &stubAuth{userID: "u1", ok: true}

	body, _ := json.Marshal(map[string]bool{"isPrivate": true})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/profile/privacy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.TogglePrivacy(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestTogglePrivacy_Unauthorized(t *testing.T) {
	h := newTestHandler()
	withDefaults(h)
	h.auth = &stubAuth{ok: false}

	body, _ := json.Marshal(map[string]bool{"isPrivate": true})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/profile/privacy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.TogglePrivacy(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestTogglePrivacy_InternalError(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.togglePrivacy = &stubTogglePrivacy{err: errors.New("db down")}
	})
	withDefaults(h)
	h.auth = &stubAuth{userID: "u1", ok: true}

	body, _ := json.Marshal(map[string]bool{"isPrivate": true})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/profile/privacy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.TogglePrivacy(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestTogglePrivacy_NotFound(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.togglePrivacy = &stubTogglePrivacy{err: user.ErrUserNotFound}
	})
	withDefaults(h)
	h.auth = &stubAuth{userID: "u1", ok: true}

	body, _ := json.Marshal(map[string]bool{"isPrivate": true})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/profile/privacy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.TogglePrivacy(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}
