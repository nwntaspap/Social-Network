package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
				IsFollowing:    true,
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

	var resp struct {
		Data struct {
			IsFollowing bool `json:"isFollowing"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Data.IsFollowing {
		t.Error("isFollowing = false, want true")
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
					{ID: "u1", Nickname: "alice", IsOnline: true},
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

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	payload, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data envelope = %T, want map", body["data"])
	}
	data, ok := payload["data"].([]any)
	if !ok || len(data) != 2 {
		t.Fatalf("users = %v, want 2 users", payload["data"])
	}
	first, ok := data[0].(map[string]any)
	if !ok {
		t.Fatalf("user[0] = %T, want map", data[0])
	}
	if first["isOnline"] != true {
		t.Errorf("isOnline for alice = %v, want true", first["isOnline"])
	}
	second, ok := data[1].(map[string]any)
	if !ok {
		t.Fatalf("user[1] = %T, want map", data[1])
	}
	if second["isOnline"] != false {
		t.Errorf("isOnline for bob = %v, want false", second["isOnline"])
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

func TestListUsers_ForwardsQueryAndPage(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.listUsers = &stubListUsers{
			result: &queries.ListUsersResult{Users: []user.User{}, Total: 0},
		}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users?query=ali&page=3", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	stub, ok := h.listUsers.(*stubListUsers)
	if !ok {
		t.Fatal("listUsers is not a *stubListUsers")
	}
	if stub.got.Query != "ali" {
		t.Errorf("Query = %q, want %q", stub.got.Query, "ali")
	}
	if stub.got.Page != 3 {
		t.Errorf("Page = %d, want 3", stub.got.Page)
	}
	if stub.got.Limit != 10 {
		t.Errorf("Limit = %d, want 10", stub.got.Limit)
	}
}

func TestListUsers_ForwardsExcludeFollowedOf(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.listUsers = &stubListUsers{
			result: &queries.ListUsersResult{Users: []user.User{}, Total: 0},
		}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users?excludeFollowedOf=viewer1", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	stub, ok := h.listUsers.(*stubListUsers)
	if !ok {
		t.Fatal("listUsers is not a *stubListUsers")
	}
	if stub.got.ExcludeUserID != "viewer1" {
		t.Errorf("ExcludeUserID = %q, want %q", stub.got.ExcludeUserID, "viewer1")
	}
}

func TestListUsers_MatchesFrontendPaginatedResponse(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	h := newTestHandler(func(h *Handler) {
		h.listUsers = &stubListUsers{
			result: &queries.ListUsersResult{
				Users: []user.User{{
					ID: "u1", Email: "alice@example.com", Nickname: "alice",
					FirstName: "Alice", LastName: "Smith",
					DateOfBirth: time.Date(1995, 3, 2, 0, 0, 0, 0, time.UTC),
					AboutMe:     "hello", IsPrivate: true, CreatedAt: now,
				}},
				Total: 25,
			},
		}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users?query=ali&page=3", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if _, ok := body["info"]; ok {
		t.Error("info envelope should not be present")
	}

	payload, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data = %#v, want unwrapped PaginatedResponse object", body["data"])
	}
	if got, ok2 := payload["page"].(float64); !ok2 || got != 3 {
		t.Errorf("page = %#v, want 3", payload["page"])
	}
	if got, ok2 := payload["pageSize"].(float64); !ok2 || got != 10 {
		t.Errorf("pageSize = %#v, want 10", payload["pageSize"])
	}
	if got, ok2 := payload["totalCount"].(float64); !ok2 || got != 25 {
		t.Errorf("totalCount = %#v, want 25", payload["totalCount"])
	}
	if got, ok2 := payload["totalPages"].(float64); !ok2 || got != 3 {
		t.Errorf("totalPages = %#v, want 3", payload["totalPages"])
	}

	data, ok := payload["data"].([]any)
	if !ok || len(data) != 1 {
		t.Fatalf("data = %#v, want array of 1", payload["data"])
	}
	u, ok := data[0].(map[string]any)
	if !ok {
		t.Fatalf("data[0] = %#v, want object", data[0])
	}
	for key, want := range map[string]any{
		"id":          "u1",
		"email":       "alice@example.com",
		"username":    "alice",
		"firstName":   "Alice",
		"lastName":    "Smith",
		"dateOfBirth": "1995-03-02",
		"aboutMe":     "hello",
		"isPublic":    false,
	} {
		if got, ok2 := u[key].(string); ok2 && got != want {
			t.Errorf("%s = %#v, want %q", key, u[key], want)
		}
	}
	if u["isPublic"] != false {
		t.Errorf("isPublic = %#v, want false", u["isPublic"])
	}
	if _, ok := u["createdAt"].(string); !ok {
		t.Errorf("createdAt = %#v, want string", u["createdAt"])
	}
}

func TestListUsers_MethodNotAllowed(t *testing.T) {
	h := newTestHandler(func(h *Handler) {
		h.listUsers = &stubListUsers{result: &queries.ListUsersResult{}}
	})
	withDefaults(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/users", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}
