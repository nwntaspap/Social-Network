package comment

import (
	"context"
	"testing"
)

// Verify a mock implementation satisfies the Repository interface at compile time.
var _ Repository = (*mockRepository)(nil)

type mockRepository struct{}

func (m *mockRepository) CreateComment(_ context.Context, _ *Comment) error      { return nil }
func (m *mockRepository) UpdateComment(_ context.Context, _ *Comment) error      { return nil }
func (m *mockRepository) DeleteComment(_ context.Context, _ string, _ int) error { return nil }
func (m *mockRepository) GetCommentByID(_ context.Context, _ int) (*Comment, error) {
	return nil, ErrCommentNotFound
}

func (m *mockRepository) GetCommentByIDWithVotes(_ context.Context, _ int, _ *string) (*Comment, error) {
	return nil, ErrCommentNotFound
}

func (m *mockRepository) GetCommentsByTopicID(_ context.Context, _ int) ([]Comment, error) {
	return nil, nil
}

func (m *mockRepository) GetCommentsByTopicIDWithVotes(_ context.Context, _ int, _ *string) ([]Comment, error) {
	return nil, nil
}

func (m *mockRepository) CastCommentVote(_ context.Context, _ string, _ int, _ int) error {
	return nil
}

func (m *mockRepository) GetVoteCounts(_ context.Context, _ int) (*VoteCounts, error) {
	return &VoteCounts{}, nil
}

func TestCommentStruct_Fields(t *testing.T) {
	c := Comment{
		ID:        1,
		TopicID:   42,
		UserID:    "user-123",
		Content:   "hello world",
		ImagePath: "/uploads/img.png",
	}

	if c.ID != 1 {
		t.Errorf("ID = %d, want 1", c.ID)
	}
	if c.TopicID != 42 {
		t.Errorf("TopicID = %d, want 42", c.TopicID)
	}
	if c.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", c.UserID, "user-123")
	}
	if c.Content != "hello world" {
		t.Errorf("Content = %q, want %q", c.Content, "hello world")
	}
	if c.ImagePath != "/uploads/img.png" {
		t.Errorf("ImagePath = %q, want %q", c.ImagePath, "/uploads/img.png")
	}
}

func TestErrCommentNotFound(t *testing.T) {
	if ErrCommentNotFound.Error() != "comment not found" {
		t.Errorf("ErrCommentNotFound = %q, want %q", ErrCommentNotFound.Error(), "comment not found")
	}
}
