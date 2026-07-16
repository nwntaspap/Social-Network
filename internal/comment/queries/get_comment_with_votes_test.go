package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
)

func TestGetCommentByIDWithVotesResolver_Success(t *testing.T) {
	expected := &comment.Comment{ID: 1, UserID: "u1", Content: "hello", UpvoteCount: 5}
	repo := &mockRepo{getByIDWithVotesResult: expected}
	r := NewGetCommentByIDWithVotesResolver(repo)

	result, err := r.Resolve(context.Background(), GetCommentByIDWithVotesQuery{CommentID: 1, UserID: "u2"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.UpvoteCount != 5 {
		t.Errorf("UpvoteCount = %d, want 5", result.UpvoteCount)
	}
}

func TestGetCommentByIDWithVotesResolver_NotFound(t *testing.T) {
	repo := &mockRepo{getByIDWithVotesErr: comment.ErrCommentNotFound}
	r := NewGetCommentByIDWithVotesResolver(repo)

	_, err := r.Resolve(context.Background(), GetCommentByIDWithVotesQuery{CommentID: 999, UserID: "u2"})
	if !errors.Is(err, comment.ErrCommentNotFound) {
		t.Errorf("error = %v, want ErrCommentNotFound", err)
	}
}

func TestGetCommentByIDWithVotesResolver_Error(t *testing.T) {
	repo := &mockRepo{getByIDWithVotesErr: errors.New("db down")}
	r := NewGetCommentByIDWithVotesResolver(repo)

	_, err := r.Resolve(context.Background(), GetCommentByIDWithVotesQuery{CommentID: 1, UserID: "u2"})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
