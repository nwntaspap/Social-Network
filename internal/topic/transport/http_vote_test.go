package transport

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"social-network/internal/topic"
)

func TestCastVote_Like(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockCastVote{})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	body := bytes.NewBufferString(`{"reactionType":1}`)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/posts/vote?id=1", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCastVote_Dislike(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockCastVote{})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	body := bytes.NewBufferString(`{"reactionType":-1}`)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/posts/vote?id=1", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCastVote_InvalidReaction(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockCastVote{err: topic.ErrInvalidVoteValue})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	body := bytes.NewBufferString(`{"reactionType":2}`)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/posts/vote?id=1", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCastVote_TopicNotFound(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockCastVote{err: topic.ErrTopicNotFound})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	body := bytes.NewBufferString(`{"reactionType":1}`)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/posts/vote?id=999", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetVoteCounts_NotFound(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockGetVotes{}, &mockGetTopic{err: topic.ErrTopicNotFound})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodGet, "/api/posts/votes/counts?id=1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteVote_NoVote(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockDeleteVote{err: topic.ErrTopicNotFound})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodDelete, "/api/posts/vote?id=1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDeleteVote_Success(t *testing.T) {
	h := newTestHandler(extractUserOK, &mockDeleteVote{})
	srv := httptest.NewServer(mux(h))
	defer srv.Close()

	resp := doReq(t, srv, http.MethodDelete, "/api/posts/vote?id=1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
