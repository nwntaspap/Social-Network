package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
)

func TestGetCommentByIDResolver_Success(t *testing.T) {
	expected := &comment.Comment{ID: 1, UserID: "u1", Content: "hello"}
	repo := &mockRepo{getByIDResult: expected}
	r := NewGetCommentByIDResolver(repo)

	result, err := r.Resolve(context.Background(), GetCommentByIDQuery{CommentID: 1})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.Content != "hello" {
		t.Errorf("Content = %q, want %q", result.Content, "hello")
	}
}

func TestGetCommentByIDResolver_NotFound(t *testing.T) {
	repo := &mockRepo{getByIDErr: comment.ErrCommentNotFound}
	r := NewGetCommentByIDResolver(repo)

	_, err := r.Resolve(context.Background(), GetCommentByIDQuery{CommentID: 999})
	if !errors.Is(err, comment.ErrCommentNotFound) {
		t.Errorf("error = %v, want ErrCommentNotFound", err)
	}
}

func TestGetCommentByIDResolver_Error(t *testing.T) {
	repo := &mockRepo{getByIDErr: errors.New("db down")}
	r := NewGetCommentByIDResolver(repo)

	_, err := r.Resolve(context.Background(), GetCommentByIDQuery{CommentID: 1})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
