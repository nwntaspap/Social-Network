package commands

import (
	"context"
	"errors"
	"testing"
)

func TestUpdateComment_Success(t *testing.T) {
	repo := &mockRepo{}
	h := NewUpdateCommentHandler(repo)

	if err := h.Execute(context.Background(), UpdateCommentCommand{
		UserID:    "u1",
		CommentID: 1,
		Content:   "updated",
	}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestUpdateComment_EmptyUserID(t *testing.T) {
	h := NewUpdateCommentHandler(&mockRepo{})
	err := h.Execute(context.Background(), UpdateCommentCommand{CommentID: 1, Content: "x"})
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("error = %v, want ErrEmptyUserID", err)
	}
}

func TestUpdateComment_ZeroCommentID(t *testing.T) {
	h := NewUpdateCommentHandler(&mockRepo{})
	err := h.Execute(context.Background(), UpdateCommentCommand{UserID: "u1", Content: "x"})
	if !errors.Is(err, ErrEmptyCommentID) {
		t.Errorf("error = %v, want ErrEmptyCommentID", err)
	}
}

func TestUpdateComment_EmptyContent(t *testing.T) {
	h := NewUpdateCommentHandler(&mockRepo{})
	err := h.Execute(context.Background(), UpdateCommentCommand{UserID: "u1", CommentID: 1})
	if !errors.Is(err, ErrEmptyContent) {
		t.Errorf("error = %v, want ErrEmptyContent", err)
	}
}

func TestUpdateComment_RepoError(t *testing.T) {
	repo := &mockRepo{updateErr: errors.New("db fail")}
	h := NewUpdateCommentHandler(repo)

	err := h.Execute(context.Background(), UpdateCommentCommand{
		UserID:    "u1",
		CommentID: 1,
		Content:   "x",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
