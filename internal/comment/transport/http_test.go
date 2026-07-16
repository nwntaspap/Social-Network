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

	"social-network/internal/comment"
	"social-network/internal/comment/commands"
	"social-network/internal/comment/queries"
)

type mockCreateComment struct {
	result *comment.Comment
	err    error
}

func (m *mockCreateComment) Execute(_ context.Context, cmd commands.CreateCommentCommand) (*comment.Comment, error) {
	return m.result, m.err
}

type mockUpdateComment struct {
	err error
}

func (m *mockUpdateComment) Execute(_ context.Context, cmd commands.UpdateCommentCommand) error {
	return m.err
}

type mockDeleteComment struct {
	err error
}

func (m *mockDeleteComment) Execute(_ context.Context, cmd commands.DeleteCommentCommand) error {
	return m.err
}

type mockGetComment struct {
	result *comment.Comment
	err    error
}

func (m *mockGetComment) Resolve(_ context.Context, q queries.GetCommentByIDQuery) (*comment.Comment, error) {
	return m.result, m.err
}

type mockGetByTopic struct {
	results []comment.Comment
	err     error
}

func (m *mockGetByTopic) Resolve(_ context.Context, q queries.GetCommentsByTopicQuery) ([]comment.Comment, error) {
	return m.results, m.err
}

func fixedTime() time.Time {
	return time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
}

func testComment() *comment.Comment {
	return &comment.Comment{
		ID:            1,
		UserID:        "user-1",
		TopicID:       10,
		Content:       "test content",
		ImagePath:     "",
		CreatedAt:     fixedTime(),
		UpdatedAt:     fixedTime(),
		UpvoteCount:   3,
		DownvoteCount: 1,
		VoteScore:     2,
		UserVote:      nil,
	}
}

func extractUserOK(_ *http.Request) (string, bool) {
	return "user-1", true
}

func extractUserFail(_ *http.Request) (string, bool) {
	return "", false
}

func handler(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateComment(w, r)
		case http.MethodPut:
			h.UpdateComment(w, r)
		case http.MethodDelete:
			h.DeleteComment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/comments/get", h.GetCommentByID)
	mux.HandleFunc("/api/comments/topic", h.GetCommentsByTopic)
	return mux
}

func doRequest(t *testing.T, srv *httptest.Server, method, path string, body any) *http.Response {
	t.Helper()

	var reqBody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, srv.URL+path, reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func decodeResponse(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return result
}

func getData(t *testing.T, result map[string]any) map[string]any {
	t.Helper()
	data, ok := result["data"]
	if !ok {
		t.Fatal("response missing 'data' key")
	}
	dataMap, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be map[string]any, got %T", data)
	}
	return dataMap
}

func getFloat(t *testing.T, m map[string]any, key string) float64 {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Fatalf("key %q not found", key)
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("expected %q to be float64, got %T", key, v)
	}
	return f
}

func TestCreateComment_Success(t *testing.T) {
	mock := &mockCreateComment{result: testComment(), err: nil}
	h := NewHandler(extractUserOK, mock, nil, nil, nil, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/comments", map[string]any{
		"topicId": 10,
		"content": "Hello world",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	result := decodeResponse(t, resp)
	data := getData(t, result)
	if data["userId"] != "user-1" {
		t.Errorf("expected userId user-1, got %v", data["userId"])
	}
	if data["content"] != "test content" {
		t.Errorf("expected content test content, got %v", data["content"])
	}
}

func TestCreateComment_Unauthorized(t *testing.T) {
	mock := &mockCreateComment{result: testComment(), err: nil}
	h := NewHandler(extractUserFail, mock, nil, nil, nil, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/comments", map[string]any{
		"topicId": 10,
		"content": "Hello world",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreateComment_BadPayload(t *testing.T) {
	mock := &mockCreateComment{result: testComment(), err: nil}
	h := NewHandler(extractUserOK, mock, nil, nil, nil, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/comments", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateComment_HandlerError(t *testing.T) {
	mock := &mockCreateComment{result: nil, err: errors.New("topic not found")}
	h := NewHandler(extractUserOK, mock, nil, nil, nil, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/comments", map[string]any{
		"topicId": 10,
		"content": "Hello world",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestUpdateComment_Success(t *testing.T) {
	mock := &mockUpdateComment{err: nil}
	h := NewHandler(extractUserOK, nil, mock, nil, nil, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPut, "/api/comments", map[string]any{
		"commentId": 1,
		"content":   "updated content",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUpdateComment_Unauthorized(t *testing.T) {
	mock := &mockUpdateComment{err: nil}
	h := NewHandler(extractUserFail, nil, mock, nil, nil, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPut, "/api/comments", map[string]any{
		"commentId": 1,
		"content":   "updated",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestDeleteComment_Success(t *testing.T) {
	mock := &mockDeleteComment{err: nil}
	h := NewHandler(extractUserOK, nil, nil, mock, nil, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodDelete, "/api/comments?id=1", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDeleteComment_BadID(t *testing.T) {
	mock := &mockDeleteComment{err: nil}
	h := NewHandler(extractUserOK, nil, nil, mock, nil, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodDelete, "/api/comments?id=abc", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetCommentByID_Success(t *testing.T) {
	mock := &mockGetComment{result: testComment(), err: nil}
	h := NewHandler(extractUserOK, nil, nil, nil, mock, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/get?id=1", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	result := decodeResponse(t, resp)
	data := getData(t, result)
	if getFloat(t, data, "id") != 1 {
		t.Errorf("expected id 1, got %v", data["id"])
	}
}

func TestGetCommentByID_NotFound(t *testing.T) {
	mock := &mockGetComment{result: nil, err: comment.ErrCommentNotFound}
	h := NewHandler(extractUserOK, nil, nil, nil, mock, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/get?id=999", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestGetCommentsByTopic_Success(t *testing.T) {
	mock := &mockGetByTopic{
		results: []comment.Comment{
			*testComment(),
			*testComment(),
		},
		err: nil,
	}
	h := NewHandler(extractUserOK, nil, nil, nil, nil, mock)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/topic?topicId=10", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	result := decodeResponse(t, resp)
	data, ok := result["data"]
	if !ok {
		t.Fatal("response missing 'data' key")
	}
	dataSlice, ok := data.([]any)
	if !ok {
		t.Fatalf("expected data to be []any, got %T", data)
	}
	if len(dataSlice) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(dataSlice))
	}
}

func TestGetCommentsByTopic_Empty(t *testing.T) {
	mock := &mockGetByTopic{results: []comment.Comment{}, err: nil}
	h := NewHandler(extractUserOK, nil, nil, nil, nil, mock)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/topic?topicId=10", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetCommentsByTopic_BadTopicID(t *testing.T) {
	mock := &mockGetByTopic{results: []comment.Comment{}, err: nil}
	h := NewHandler(extractUserOK, nil, nil, nil, nil, mock)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/topic?topicId=abc", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
