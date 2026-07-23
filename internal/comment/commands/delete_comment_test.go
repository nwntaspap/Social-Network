package commands

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteComment_Success(t *testing.T) {
	repo := &mockRepo{}
	h := NewDeleteCommentHandler(repo)

	if err := h.Execute(context.Background(), DeleteCommentCommand{
		UserID:    "u1",
		CommentID: 1,
	}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestDeleteComment_EmptyUserID(t *testing.T) {
	h := NewDeleteCommentHandler(&mockRepo{})
	err := h.Execute(context.Background(), DeleteCommentCommand{CommentID: 1})
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("error = %v, want ErrEmptyUserID", err)
	}
}

func TestDeleteComment_ZeroCommentID(t *testing.T) {
	h := NewDeleteCommentHandler(&mockRepo{})
	err := h.Execute(context.Background(), DeleteCommentCommand{UserID: "u1"})
	if !errors.Is(err, ErrEmptyCommentID) {
		t.Errorf("error = %v, want ErrEmptyCommentID", err)
	}
}

func TestDeleteComment_RepoError(t *testing.T) {
	repo := &mockRepo{deleteErr: errors.New("db fail")}
	h := NewDeleteCommentHandler(repo)

	err := h.Execute(context.Background(), DeleteCommentCommand{
		UserID:    "u1",
		CommentID: 1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
