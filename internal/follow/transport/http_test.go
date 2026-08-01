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

	"social-network/internal/follow"
)

func TestFollowUser_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{err: errors.New("db error")}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow", nil)
	w := httptest.NewRecorder()

	h.FollowUser(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestUnfollowUser_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	body, _ := json.Marshal(map[string]string{"targetId": "user-2"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow/unfollow", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UnfollowUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUnfollowUser_Unauthorized(t *testing.T) {
	h := NewHandler(testExtractor(""),
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow/unfollow", nil)
	w := httptest.NewRecorder()

	h.UnfollowUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUnfollowUser_CommandError(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{err: errors.New("db error")}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	body, _ := json.Marshal(map[string]string{"targetId": "user-2"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow/unfollow", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UnfollowUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestUnfollowUser_InvalidMethod(t *testing.T) {
	h := NewHandler(testExtractor("user-1"),
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/unfollow", nil)
	w := httptest.NewRecorder()

	h.UnfollowUser(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestAcceptRequest_Success(t *testing.T) {
	h := NewHandler(testExtractor("user-2"),
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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

func TestGetPendingRequests_MatchesFrontendFollowRequest(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	h := NewHandler(testExtractor("user-1"),
		&mockUserLookup{byID: map[string]*UserResult{
			"user-3": {ID: "user-3", Username: "alice", FirstName: "Alice"},
			"user-1": {ID: "user-1", Username: "bob", FirstName: "Bob"},
		}},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{},
		&mockGetPendingRequests{result: []follow.Request{{
			FollowerID: "user-3",
			FolloweeID: "user-1",
			CreatedAt:  now,
		}}},
		&mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/requests", nil)
	w := httptest.NewRecorder()

	h.GetPendingRequests(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(body.Data))
	}
	d := body.Data[0]

	if got, ok := d["id"].(string); !ok || got != "user-3:user-1" {
		t.Errorf("id = %#v, want %q", d["id"], "user-3:user-1")
	}
	if got, ok := d["requesterId"].(string); !ok || got != "user-3" {
		t.Errorf("requesterId = %#v, want %q", d["requesterId"], "user-3")
	}
	if got, ok := d["targetId"].(string); !ok || got != "user-1" {
		t.Errorf("targetId = %#v, want %q", d["targetId"], "user-1")
	}
	if got, ok := d["status"].(string); !ok || got != "pending" {
		t.Errorf("status = %#v, want %q", d["status"], "pending")
	}
	if _, ok := d["createdAt"].(string); !ok {
		t.Errorf("createdAt = %#v, want string", d["createdAt"])
	}

	requester, ok := d["requester"].(map[string]any)
	if !ok {
		t.Fatalf("requester = %#v, want object", d["requester"])
	}
	if got, ok2 := requester["username"].(string); !ok2 || got != "alice" {
		t.Errorf("requester.username = %#v, want %q", requester["username"], "alice")
	}
	if got, ok2 := requester["firstName"].(string); !ok2 || got != "Alice" {
		t.Errorf("requester.firstName = %#v, want %q", requester["firstName"], "Alice")
	}

	target, ok := d["target"].(map[string]any)
	if !ok {
		t.Fatalf("target = %#v, want object", d["target"])
	}
	if got, ok2 := target["username"].(string); !ok2 || got != "bob" {
		t.Errorf("target.username = %#v, want %q", target["username"], "bob")
	}
}

func TestGetPendingRequests_Unauthorized(t *testing.T) {
	h := NewHandler(testExtractor(""),
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
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
		&mockUserLookup{},
		&mockFollowUser{}, &mockUnfollowUser{}, &mockAcceptRequest{}, &mockDeclineRequest{},
		&mockGetFollowers{}, &mockGetFollowing{}, &mockGetPendingRequests{}, &mockAreConnected{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/connected?targetId=user-2", nil)
	w := httptest.NewRecorder()

	h.AreConnected(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
