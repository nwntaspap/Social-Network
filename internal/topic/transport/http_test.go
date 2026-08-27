package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-network/internal/platform/logger"
	"social-network/internal/topic"
	"social-network/internal/topic/commands"
	"social-network/internal/topic/queries"
)

type mockCreateTopic struct {
	result *topic.Topic
	err    error
}

func (m *mockCreateTopic) Execute(_ context.Context, cmd commands.CreateTopicCommand) (*topic.Topic, error) {
	return m.result, m.err
}

type mockUpdateTopic struct {
	result *topic.Topic
	err    error
}

func (m *mockUpdateTopic) Execute(_ context.Context, cmd commands.UpdateTopicCommand) (*topic.Topic, error) {
	return m.result, m.err
}

type mockDeleteTopic struct {
	err error
}

func (m *mockDeleteTopic) Execute(_ context.Context, cmd commands.DeleteTopicCommand) error {
	return m.err
}

type mockCastVote struct {
	err error
}

func (m *mockCastVote) Execute(_ context.Context, cmd commands.CastVoteCommand) error {
	return m.err
}

type mockDeleteVote struct {
	err error
}

func (m *mockDeleteVote) Execute(_ context.Context, cmd commands.DeleteVoteCommand) error {
	return m.err
}

type mockGetFeed struct {
	result *queries.GetFeedResult
	err    error
}

func (m *mockGetFeed) Resolve(_ context.Context, q queries.GetFeedQuery) (*queries.GetFeedResult, error) {
	return m.result, m.err
}

type mockGetTopic struct {
	result *topic.Topic
	err    error
}

func (m *mockGetTopic) Resolve(_ context.Context, q queries.GetTopicQuery) (*topic.Topic, error) {
	return m.result, m.err
}

type mockGetByUser struct {
	result *queries.GetTopicsByUserResult
	err    error
}

func (m *mockGetByUser) Resolve(_ context.Context, q queries.GetTopicsByUserQuery) (*queries.GetTopicsByUserResult, error) {
	return m.result, m.err
}

type mockGetByGroup struct {
	result *queries.GetTopicsByGroupResult
	err    error
}

func (m *mockGetByGroup) Resolve(_ context.Context, q queries.GetTopicsByGroupQuery) (*queries.GetTopicsByGroupResult, error) {
	return m.result, m.err
}

type mockGetVotes struct {
	result *topic.VoteCounts
	err    error
}

func (m *mockGetVotes) Resolve(_ context.Context, q queries.GetVoteCountsQuery) (*topic.VoteCounts, error) {
	return m.result, m.err
}

type mockUserLookup struct {
	result *UserResult
	err    error
}

func (m *mockUserLookup) GetUserByID(_ context.Context, id string) (*UserResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &UserResult{ID: id, Username: "alice", Nickname: "alice"}, nil
}

func fixedTime() time.Time {
	return time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
}

func testTopic() *topic.Topic {
	return &topic.Topic{
		ID:            1,
		UserID:        "user-1",
		Title:         "Test Post",
		Content:       "Test content",
		ImagePath:     "/images/photo.jpg",
		Visibility:    topic.VisibilityPublic,
		CreatedAt:     fixedTime(),
		UpdatedAt:     fixedTime(),
		UpvoteCount:   3,
		DownvoteCount: 1,
		VoteScore:     2,
		CommentsCount: 2,
	}
}

func extractUserOK(_ *http.Request) (string, bool) {
	return "user-1", true
}

func extractUserFail(_ *http.Request) (string, bool) {
	return "", false
}

func mux(h *Handler) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateTopic(w, r)
		case http.MethodPut:
			h.UpdateTopic(w, r)
		case http.MethodDelete:
			h.DeleteTopic(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	m.HandleFunc("/api/posts/feed", h.GetFeed)
	m.HandleFunc("/api/posts/vote", h.CastVote)
	m.HandleFunc("/api/posts/vote-counts", h.GetVoteCounts)
	m.HandleFunc("/api/posts/get", h.GetTopic)
	m.HandleFunc("/api/posts/user", h.GetUserTopics)
	m.HandleFunc("/api/posts/group", h.GetGroupTopics)
	return m
}

func doReq(t *testing.T, srv *httptest.Server, method, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

func newTestHandler(extractor UserExtractor, mocks ...any) *Handler {
	var create CreateTopicExecutor
	var update UpdateTopicExecutor
	var del DeleteTopicExecutor
	var cast CastVoteExecutor
	var delVote DeleteVoteExecutor
	var getFeed GetFeedResolver
	var getTopic GetTopicResolver
	var getByUser GetTopicsByUserResolver
	var getByGroup GetTopicsByGroupResolver
	var getVotes GetVoteCountsResolver
	var lookup UserLookup

	for _, m := range mocks {
		switch v := m.(type) {
		case *mockCreateTopic:
			create = v
		case *mockUpdateTopic:
			update = v
		case *mockDeleteTopic:
			del = v
		case *mockCastVote:
			cast = v
		case *mockDeleteVote:
			delVote = v
		case *mockGetFeed:
			getFeed = v
		case *mockGetTopic:
			getTopic = v
		case *mockGetByUser:
			getByUser = v
		case *mockGetByGroup:
			getByGroup = v
		case *mockGetVotes:
			getVotes = v
		case *mockUserLookup:
			lookup = v
		}
	}

	if lookup == nil {
		lookup = &mockUserLookup{}
	}

	return NewHandler(extractor, lookup, create, update, del, cast, delVote, getFeed, getTopic, getByUser, getByGroup, getVotes,
		logger.New(io.Discard, logger.LevelOff))
}

func doMultipartReq(t *testing.T, srv *httptest.Server, method, path string, fields map[string]string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	_ = w.Close()

	req, err := http.NewRequestWithContext(context.Background(), method, srv.URL+path, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

func TestCreateTopic_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockCreateTopic{result: testTopic()})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doMultipartReq(t, srv, http.MethodPost, "/api/posts", map[string]string{
		"title":   "Test Post",
		"content": "Test content",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestCreateTopic_Unauthorized(t *testing.T) {
	h := newTestHandler(extractUserFail, &mockCreateTopic{result: testTopic()})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doMultipartReq(t, srv, http.MethodPost, "/api/posts", map[string]string{
		"title":   "Test",
		"content": "Test",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestUpdateTopic_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockUpdateTopic{result: testTopic()})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doMultipartReq(t, srv, http.MethodPut, "/api/posts?id=1", map[string]string{
		"title":   "Updated",
		"content": "New content",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDeleteTopic_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockDeleteTopic{})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodDelete, "/api/posts?id=1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDeleteTopic_BadID(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockDeleteTopic{})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodDelete, "/api/posts?id=abc")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetFeed_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetFeed{
		result: &queries.GetFeedResult{
			Topics: []topic.Topic{*testTopic()},
			Total:  1,
		},
	})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodGet, "/api/posts/feed?page=1&limit=10")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetTopic_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetTopic{result: testTopic()})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodGet, "/api/posts/get?id=1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetTopic_NotFound(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetTopic{err: topic.ErrTopicNotFound})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodGet, "/api/posts/get?id=1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetTopic_Error(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetTopic{err: errors.New("not found")})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodGet, "/api/posts/get?id=1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestTopicResponse_MatchesFrontendPost(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetTopic{result: testTopic()})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodGet, "/api/posts/get?id=1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	d := body.Data

	if got, ok := d["id"].(string); !ok || got != "1" {
		t.Errorf("id = %#v, want string \"1\"", d["id"])
	}
	if _, ok := d["user"].(map[string]any); !ok {
		t.Errorf("user = %#v, want nested object", d["user"])
	}
	if got, ok := d["imageUrl"].(string); !ok || got != "/images/photo.jpg" {
		t.Errorf("imageUrl = %#v, want %q", d["imageUrl"], "/images/photo.jpg")
	}
	if got, ok := d["commentsCount"].(float64); !ok || got != 2 {
		t.Errorf("commentsCount = %#v, want 2", d["commentsCount"])
	}
	if got, ok := d["likesCount"].(float64); !ok || got != 3 {
		t.Errorf("likesCount = %#v, want 3", d["likesCount"])
	}
	if got, ok := d["downvotesCount"].(float64); !ok || got != 1 {
		t.Errorf("downvotesCount = %#v, want 1", d["downvotesCount"])
	}
	if _, ok := d["allowedUsers"]; ok {
		t.Error("allowedUsers should be omitted for non-owner/empty list")
	}
	if _, ok := d["ownerUsername"]; ok {
		t.Error("ownerUsername should be removed from the response")
	}
	if _, ok := d["imagePath"]; ok {
		t.Error("imagePath should be renamed to imageUrl")
	}
}
