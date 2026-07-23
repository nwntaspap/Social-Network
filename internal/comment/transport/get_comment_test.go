package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/internal/comment"
)

func TestGetCommentByID_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetComment{result: testComment(), err: nil})
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
	h := newTestHandler(extractUserOK, &mockGetComment{result: nil, err: comment.ErrCommentNotFound})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/get?id=999", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestGetCommentByIDWithVotes_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetCommentWV{result: testComment(), err: nil})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/get/votes?id=1", nil)
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

func TestGetCommentByIDWithVotes_Unauthorized(t *testing.T) {
	h := newTestHandler(extractUserFail, &mockGetCommentWV{result: testComment(), err: nil})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/get/votes?id=1", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}
