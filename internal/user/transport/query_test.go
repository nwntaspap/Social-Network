package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/internal/user"
	"social-network/internal/user/queries"
)

func TestGetProfile_Success(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.getProfile = &stubGetProfile{
			result: &queries.ProfileResult{
				User:           user.User{ID: "u1", Nickname: "nick"},
				FollowerCount:  5,
				FollowingCount: 3,
			},
		}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/user/profile?user_id=u1", nil)
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestGetProfile_MissingUserID(t *testing.T) {
	h := newTestHandler()
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/user/profile", nil)
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.getProfile = &stubGetProfile{err: user.ErrUserNotFound}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/user/profile?user_id=unknown", nil)
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGetProfile_WithAuth(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.getProfile = &stubGetProfile{
			result: &queries.ProfileResult{
				User: user.User{ID: "target"},
			},
		}
	})
	h.auth = &stubAuth{userID: "viewer", ok: true}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/user/profile?user_id=target", nil)
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestGetActivity_Success(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.getActivity = &stubGetActivity{
			result: &queries.ActivityResult{
				PostCount:    10,
				CommentCount: 5,
				VoteCount:    3,
			},
		}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/user/activity?user_id=u1", nil)
	rr := httptest.NewRecorder()

	h.GetActivity(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestGetActivity_MissingUserID(t *testing.T) {
	h := newTestHandler()
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/user/activity", nil)
	rr := httptest.NewRecorder()

	h.GetActivity(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetActivity_InternalError(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.getActivity = &stubGetActivity{err: errors.New("db down")}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/user/activity?user_id=u1", nil)
	rr := httptest.NewRecorder()

	h.GetActivity(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestListUsers_Success(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.listUsers = &stubListUsers{
			result: &queries.ListUsersResult{
				Users: []user.User{
					{ID: "u1", Nickname: "alice"},
					{ID: "u2", Nickname: "bob"},
				},
			},
		}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestListUsers_InternalError(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.listUsers = &stubListUsers{err: errors.New("db down")}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}
