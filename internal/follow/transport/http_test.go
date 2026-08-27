package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/follow"
	"social-network/internal/follow/commands"
	"social-network/internal/platform/logger"
)

type testMocks struct {
	extractor UserExtractor
	lookup    *mockUserLookup
	follow    *mockFollowUser
	unfollow  *mockUnfollowUser
	accept    *mockAcceptRequest
	decline   *mockDeclineRequest
	followers *mockGetFollowers
	following *mockGetFollowing
	requests  *mockGetPendingRequests
	connected *mockAreConnected
}

func defaultMocks() *testMocks {
	return &testMocks{
		extractor: testExtractor("user-1"),
		lookup:    &mockUserLookup{},
		follow:    &mockFollowUser{},
		unfollow:  &mockUnfollowUser{},
		accept:    &mockAcceptRequest{},
		decline:   &mockDeclineRequest{},
		followers: &mockGetFollowers{},
		following: &mockGetFollowing{},
		requests:  &mockGetPendingRequests{},
		connected: &mockAreConnected{},
	}
}

func (m *testMocks) handler() *Handler {
	return NewHandler(m.extractor, m.lookup, m.follow, m.unfollow, m.accept, m.decline,
		m.followers, m.following, m.requests, m.connected,
		logger.New(io.Discard, logger.LevelOff))
}

func (m *testMocks) unauthorized() *testMocks {
	m.extractor = testExtractor("")
	return m
}

func TestFollowUser_Success(t *testing.T) {
	m := defaultMocks()
	m.follow.result = commands.FollowedDirect
	h := m.handler()

	body, _ := json.Marshal(map[string]string{"targetId": "user-2"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.FollowUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Status != "following" {
		t.Errorf("status = %q, want %q", resp.Data.Status, "following")
	}
}

func TestFollowUser_PendingStatus(t *testing.T) {
	m := defaultMocks()
	m.follow.result = commands.FollowPending
	h := m.handler()

	body, _ := json.Marshal(map[string]string{"targetId": "user-2"})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.FollowUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Status != "pending" {
		t.Errorf("status = %q, want %q", resp.Data.Status, "pending")
	}
}

func TestFollowUser_Unauthorized(t *testing.T) {
	h := defaultMocks().unauthorized().handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow", nil)
	w := httptest.NewRecorder()

	h.FollowUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestFollowUser_CommandError(t *testing.T) {
	m := defaultMocks()
	m.follow.err = errors.New("db error")
	h := m.handler()

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
	h := defaultMocks().handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow", nil)
	w := httptest.NewRecorder()

	h.FollowUser(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestUnfollowUser_Success(t *testing.T) {
	h := defaultMocks().handler()

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
	h := defaultMocks().unauthorized().handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow/unfollow", nil)
	w := httptest.NewRecorder()

	h.UnfollowUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUnfollowUser_CommandError(t *testing.T) {
	m := defaultMocks()
	m.unfollow.err = errors.New("db error")
	h := m.handler()

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
	h := defaultMocks().handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/unfollow", nil)
	w := httptest.NewRecorder()

	h.UnfollowUser(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestAcceptRequest_Success(t *testing.T) {
	m := defaultMocks()
	m.extractor = testExtractor("user-2")
	h := m.handler()

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
	h := defaultMocks().unauthorized().handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/follow/accept", nil)
	w := httptest.NewRecorder()

	h.AcceptRequest(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestDeclineRequest_Success(t *testing.T) {
	m := defaultMocks()
	m.extractor = testExtractor("user-2")
	h := m.handler()

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
	m := defaultMocks()
	m.lookup.byID = map[string]*UserResult{
		"user-2": {ID: "user-2", Username: "bob", FirstName: "Bob", LastName: "Smith", AvatarURL: "avatar-bob"},
	}
	m.followers.result = []follow.Follow{{FollowerID: "user-2", FolloweeID: "user-1"}}
	h := m.handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/followers?userId=user-1", nil)
	w := httptest.NewRecorder()

	h.GetFollowers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp struct {
		Data []*UserResult `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0] == nil {
		t.Fatalf("data = %v, want one enriched user", resp.Data)
	}
	if resp.Data[0].ID != "user-2" || resp.Data[0].Username != "bob" {
		t.Errorf("user = %+v, want id=user-2 username=bob", resp.Data[0])
	}
}

func TestGetFollowers_MissingParam(t *testing.T) {
	h := defaultMocks().handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/followers", nil)
	w := httptest.NewRecorder()

	h.GetFollowers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetFollowing_Success(t *testing.T) {
	m := defaultMocks()
	m.following.result = []follow.Follow{{FollowerID: "user-1", FolloweeID: "user-2"}}
	h := m.handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/following?userId=user-1", nil)
	w := httptest.NewRecorder()

	h.GetFollowing(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp struct {
		Data []*UserResult `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0] == nil {
		t.Fatalf("data = %v, want one enriched user", resp.Data)
	}
	if resp.Data[0].ID != "user-2" {
		t.Errorf("user id = %q, want %q", resp.Data[0].ID, "user-2")
	}
}

func TestGetPendingRequests_Success(t *testing.T) {
	m := defaultMocks()
	m.requests.result = []follow.Request{{FollowerID: "user-3", FolloweeID: "user-1"}}
	h := m.handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/requests", nil)
	w := httptest.NewRecorder()

	h.GetPendingRequests(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetPendingRequests_MatchesFrontendFollowRequest(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	m := defaultMocks()
	m.lookup.byID = map[string]*UserResult{
		"user-3": {ID: "user-3", Username: "alice", FirstName: "Alice"},
		"user-1": {ID: "user-1", Username: "bob", FirstName: "Bob"},
	}
	m.requests.result = []follow.Request{{
		FollowerID: "user-3",
		FolloweeID: "user-1",
		CreatedAt:  now,
	}}
	h := m.handler()

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
	h := defaultMocks().unauthorized().handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/requests", nil)
	w := httptest.NewRecorder()

	h.GetPendingRequests(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAreConnected_Success(t *testing.T) {
	m := defaultMocks()
	m.connected.result = true
	h := m.handler()

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
	h := defaultMocks().handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/connected", nil)
	w := httptest.NewRecorder()

	h.AreConnected(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAreConnected_Unauthorized(t *testing.T) {
	h := defaultMocks().unauthorized().handler()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/follow/connected?targetId=user-2", nil)
	w := httptest.NewRecorder()

	h.AreConnected(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
