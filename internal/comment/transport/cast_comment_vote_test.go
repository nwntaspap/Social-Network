package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/internal/comment"
)

func TestCastCommentVote_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockCastCommentVote{err: nil})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/comments/cast-vote?id=1", map[string]any{
		"reactionType": 1,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCastCommentVote_Downvote(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockCastCommentVote{err: nil})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/comments/cast-vote?id=1", map[string]any{
		"reactionType": -1,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCastCommentVote_Unauthorized(t *testing.T) {
	h := newTestHandler(extractUserFail, &mockCastCommentVote{err: nil})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/comments/cast-vote?id=1", map[string]any{
		"reactionType": 1,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCastCommentVote_BadID(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockCastCommentVote{err: nil})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/comments/cast-vote?id=abc", map[string]any{
		"reactionType": 1,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCastCommentVote_HandlerError(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockCastCommentVote{err: comment.ErrInvalidVoteValue})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/comments/cast-vote?id=1", map[string]any{
		"reactionType": 1,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestGetVoteCounts_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetCommentVotes{
		result: &comment.VoteCounts{Upvotes: 5, Downvotes: 2, Score: 3},
		err:    nil,
	})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/vote-counts?id=1", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	result := decodeResponse(t, resp)
	data := getData(t, result)
	if getFloat(t, data, "upvotes") != 5 {
		t.Errorf("expected upvotes 5, got %v", data["upvotes"])
	}
	if getFloat(t, data, "downvotes") != 2 {
		t.Errorf("expected downvotes 2, got %v", data["downvotes"])
	}
	if getFloat(t, data, "score") != 3 {
		t.Errorf("expected score 3, got %v", data["score"])
	}
}

func TestGetVoteCounts_BadID(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetCommentVotes{
		result: &comment.VoteCounts{},
		err:    nil,
	})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/vote-counts?id=abc", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetVoteCounts_Error(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetCommentVotes{
		result: nil,
		err:    comment.ErrCommentNotFound,
	})
	srv := httptest.NewServer(handler(h))
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/comments/vote-counts?id=999", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}
