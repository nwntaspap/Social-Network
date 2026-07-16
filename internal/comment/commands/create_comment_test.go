package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
)

type mockRepo struct {
	createErr error
	updateErr error
	deleteErr error
	getResult *comment.Comment
	getErr    error
}

func (m *mockRepo) CreateComment(_ context.Context, c *comment.Comment) error {
	return m.createErr
}

func (m *mockRepo) UpdateComment(_ context.Context, _ *comment.Comment) error {
	return m.updateErr
}

func (m *mockRepo) DeleteComment(_ context.Context, _ string, _ int) error {
	return m.deleteErr
}

func (m *mockRepo) GetCommentByID(_ context.Context, _ int) (*comment.Comment, error) {
	return m.getResult, m.getErr
}

func (m *mockRepo) GetCommentByIDWithVotes(_ context.Context, _ int, _ *string) (*comment.Comment, error) {
	return m.getResult, m.getErr
}

func (m *mockRepo) GetCommentsByTopicID(_ context.Context, _ int) ([]comment.Comment, error) {
	return nil, nil
}

func (m *mockRepo) GetCommentsByTopicIDWithVotes(_ context.Context, _ int, _ *string) ([]comment.Comment, error) {
	return nil, nil
}

var (
	validPNG  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	badHeader = []byte{0x00, 0x01, 0x02, 0x03}
)

func TestCreateComment_Success(t *testing.T) {
	repo := &mockRepo{}
	h := NewCreateCommentHandler(repo)

	c, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:  "u1",
		TopicID: 1,
		Content: "hello",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if c.UserID != "u1" {
		t.Errorf("UserID = %q, want %q", c.UserID, "u1")
	}
	if c.Content != "hello" {
		t.Errorf("Content = %q, want %q", c.Content, "hello")
	}
}

func TestCreateComment_WithValidImage(t *testing.T) {
	repo := &mockRepo{}
	h := NewCreateCommentHandler(repo)

	c, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:    "u1",
		TopicID:   1,
		Content:   "pic",
		ImageData: validPNG,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if c == nil {
		t.Fatal("expected comment, got nil")
	}
}

func TestCreateComment_InvalidImage(t *testing.T) {
	repo := &mockRepo{}
	h := NewCreateCommentHandler(repo)

	_, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:    "u1",
		TopicID:   1,
		Content:   "bad pic",
		ImageData: badHeader,
	})
	if err == nil {
		t.Fatal("expected error for bad image, got nil")
	}
}

func TestCreateComment_EmptyUserID(t *testing.T) {
	h := NewCreateCommentHandler(&mockRepo{})
	_, err := h.Execute(context.Background(), CreateCommentCommand{TopicID: 1, Content: "x"})
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("error = %v, want ErrEmptyUserID", err)
	}
}

func TestCreateComment_ZeroTopicID(t *testing.T) {
	h := NewCreateCommentHandler(&mockRepo{})
	_, err := h.Execute(context.Background(), CreateCommentCommand{UserID: "u1", Content: "x"})
	if !errors.Is(err, ErrEmptyTopicID) {
		t.Errorf("error = %v, want ErrEmptyTopicID", err)
	}
}

func TestCreateComment_EmptyContent(t *testing.T) {
	h := NewCreateCommentHandler(&mockRepo{})
	_, err := h.Execute(context.Background(), CreateCommentCommand{UserID: "u1", TopicID: 1})
	if !errors.Is(err, ErrEmptyContent) {
		t.Errorf("error = %v, want ErrEmptyContent", err)
	}
}

func TestCreateComment_RepoError(t *testing.T) {
	repo := &mockRepo{createErr: errors.New("db fail")}
	h := NewCreateCommentHandler(repo)

	_, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:  "u1",
		TopicID: 1,
		Content: "x",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateComment_NoImageAllowed(t *testing.T) {
	repo := &mockRepo{}
	h := NewCreateCommentHandler(repo)

	c, err := h.Execute(context.Background(), CreateCommentCommand{
		UserID:  "u1",
		TopicID: 1,
		Content: "no image",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if c.ImagePath != "" {
		t.Errorf("ImagePath = %q, want empty", c.ImagePath)
	}
}
