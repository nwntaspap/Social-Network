package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
)

func TestDeleteComment_Success(t *testing.T) {
	repo := &mockRepo{getResult: &comment.Comment{ID: 1, TopicID: 1, UserID: "u1", Content: "hello"}}
	h := NewDeleteCommentHandler(repo, &mockBus{}, &mockTopicRepo{}, &mockUserRepo{})

	if err := h.Execute(context.Background(), DeleteCommentCommand{
		UserID:    "u1",
		CommentID: 1,
	}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestDeleteComment_EmptyUserID(t *testing.T) {
	h := NewDeleteCommentHandler(&mockRepo{}, &mockBus{}, &mockTopicRepo{}, &mockUserRepo{})
	err := h.Execute(context.Background(), DeleteCommentCommand{CommentID: 1})
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("error = %v, want ErrEmptyUserID", err)
	}
}

func TestDeleteComment_ZeroCommentID(t *testing.T) {
	h := NewDeleteCommentHandler(&mockRepo{}, &mockBus{}, &mockTopicRepo{}, &mockUserRepo{})
	err := h.Execute(context.Background(), DeleteCommentCommand{UserID: "u1"})
	if !errors.Is(err, ErrEmptyCommentID) {
		t.Errorf("error = %v, want ErrEmptyCommentID", err)
	}
}

func TestDeleteComment_RepoError(t *testing.T) {
	repo := &mockRepo{deleteErr: errors.New("db fail"), getResult: &comment.Comment{ID: 1, TopicID: 1, UserID: "u1", Content: "hello"}}
	h := NewDeleteCommentHandler(repo, &mockBus{}, &mockTopicRepo{}, &mockUserRepo{})

	err := h.Execute(context.Background(), DeleteCommentCommand{
		UserID:    "u1",
		CommentID: 1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
