package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/internal/follow"
	"social-network/internal/follow/commands"
	"social-network/internal/follow/queries"
)

type mockFollowUser struct {
	err error
}

func (m *mockFollowUser) Execute(_ context.Context, _ commands.FollowUserCommand) error {
	return m.err
}

type mockAcceptRequest struct {
	err error
}

func (m *mockAcceptRequest) Execute(_ context.Context, _ commands.AcceptRequestCommand) error {
	return m.err
}

type mockDeclineRequest struct {
	err error
}

func (m *mockDeclineRequest) Execute(_ context.Context, _ commands.DeclineRequestCommand) error {
	return m.err
}

type mockGetFollowers struct {
	result []follow.Follow
	err    error
}

func (m *mockGetFollowers) Resolve(_ context.Context, _ queries.GetFollowersQuery) ([]follow.Follow, error) {
	return m.result, m.err
}

type mockGetFollowing struct {
	result []follow.Follow
	err    error
}

func (m *mockGetFollowing) Resolve(_ context.Context, _ queries.GetFollowingQuery) ([]follow.Follow, error) {
	return m.result, m.err
}

type mockGetPendingRequests struct {
	result []follow.Request
	err    error
}

func (m *mockGetPendingRequests) Resolve(_ context.Context, _ queries.GetPendingRequestsQuery) ([]follow.Request, error) {
	return m.result, m.err
}

type mockAreConnected struct {
	result bool
	err    error
}

func (m *mockAreConnected) Resolve(_ context.Context, _ queries.AreConnectedQuery) (bool, error) {
	return m.result, m.err
}

func testExtractor(userID string) UserExtractor {
	return func(_ *http.Request) (string, bool) {
		if userID == "" {
			return "", false
		}
		return userID, true
	}
}

func TestFollowUser_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	body, _ := json.Marshal(map[string]string{"targetId": "user-2"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.FollowUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestFollowUser_Unauthorized(t *testing.T) {
	h := NewHandler(testExtractor(""),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow", nil)
	w := httptest.NewRecorder()

	h.FollowUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestFollowUser_CommandError(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockFollowUser{err: errors.New("db error")}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	body, _ := json.Marshal(map[string]string{"targetId": "user-2"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.FollowUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestFollowUser_InvalidMethod(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow", nil)
	w := httptest.NewRecorder()

	h.FollowUser(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestAcceptRequest_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-2"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	body, _ := json.Marshal(map[string]string{"followerId": "user-1"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow/accept", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AcceptRequest(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAcceptRequest_Unauthorized(t *testing.T) {
	h := NewHandler(testExtractor(""),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow/accept", nil)
	w := httptest.NewRecorder()

	h.AcceptRequest(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestDeclineRequest_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-2"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	body, _ := json.Marshal(map[string]string{"followerId": "user-1"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow/decline", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.DeclineRequest(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetFollowers_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{result: []follow.Follow{{FollowerID: "user-2", FolloweeID: "user-1"}}},
		&mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/followers?userId=user-1", nil)
	w := httptest.NewRecorder()

	h.GetFollowers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetFollowers_MissingParam(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/followers", nil)
	w := httptest.NewRecorder()

	h.GetFollowers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetFollowing_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{result: []follow.Follow{{FollowerID: "user-1", FolloweeID: "user-2"}}},
		&mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/following?userId=user-1", nil)
	w := httptest.NewRecorder()

	h.GetFollowing(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetPendingRequests_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{},
		&mockGetPendingRequests{result: []follow.Request{{FollowerID: "user-3", FolloweeID: "user-1"}}},
		&mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/requests", nil)
	w := httptest.NewRecorder()

	h.GetPendingRequests(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetPendingRequests_Unauthorized(t *testing.T) {
	h := NewHandler(testExtractor(""),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/requests", nil)
	w := httptest.NewRecorder()

	h.GetPendingRequests(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAreConnected_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{},
		&mockAreConnected{result: true})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/connected?targetId=user-2", nil)
	w := httptest.NewRecorder()

	h.AreConnected(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Data map[string]bool `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Data["connected"] {
		t.Error("connected = false, want true")
	}
}

func TestAreConnected_MissingParam(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/connected", nil)
	w := httptest.NewRecorder()

	h.AreConnected(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAreConnected_Unauthorized(t *testing.T) {
	h := NewHandler(testExtractor(""),
		&mockFollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/connected?targetId=user-2", nil)
	w := httptest.NewRecorder()

	h.AreConnected(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
