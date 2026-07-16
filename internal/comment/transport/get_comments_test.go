package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/internal/comment"
)

func TestGetCommentsByTopic_Success(t *testing.T) {
	mock := &mockGetByTopic{
		results: []comment.Comment{
			*testComment(),
			*testComment(),
		},
		err: nil,
	}
	h := NewHandler(extractUserOK, nil, nil, nil, nil, nil, mock, nil)
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
	h := NewHandler(extractUserOK, nil, nil, nil, nil, nil, mock, nil)
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
	h := NewHandler(extractUserOK, nil, nil, nil, nil, nil, mock, nil)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/topic?topicId=abc", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetCommentsByTopicWithVotes_Success(t *testing.T) {
	mock := &mockGetByTopicWV{
		results: []comment.Comment{
			*testComment(),
			*testComment(),
		},
		err: nil,
	}
	h := NewHandler(extractUserOK, nil, nil, nil, nil, nil, nil, mock)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/topic/votes?topicId=10", nil)
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

func TestGetCommentsByTopicWithVotes_Unauthorized(t *testing.T) {
	mock := &mockGetByTopicWV{results: []comment.Comment{}, err: nil}
	h := NewHandler(extractUserFail, nil, nil, nil, nil, nil, nil, mock)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/topic/votes?topicId=10", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetCommentsByTopicWithVotes_Empty(t *testing.T) {
	mock := &mockGetByTopicWV{results: []comment.Comment{}, err: nil}
	h := NewHandler(extractUserOK, nil, nil, nil, nil, nil, nil, mock)
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/topic/votes?topicId=10", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
