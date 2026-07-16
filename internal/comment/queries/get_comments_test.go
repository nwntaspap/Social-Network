package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
)

type mockRepo struct {
	getResult   []comment.Comment
	getErr      error
	getResultWV []comment.Comment
	getErrWV    error
}

func (m *mockRepo) CreateComment(_ context.Context, _ *comment.Comment) error { return nil }
func (m *mockRepo) UpdateComment(_ context.Context, _ *comment.Comment) error { return nil }
func (m *mockRepo) DeleteComment(_ context.Context, _ string, _ int) error    { return nil }
func (m *mockRepo) GetCommentByID(_ context.Context, _ int) (*comment.Comment, error) {
	return nil, comment.ErrCommentNotFound
}

func (m *mockRepo) GetCommentByIDWithVotes(_ context.Context, _ int, _ *string) (*comment.Comment, error) {
	return nil, comment.ErrCommentNotFound
}

func (m *mockRepo) GetCommentsByTopicID(_ context.Context, _ int) ([]comment.Comment, error) {
	return m.getResult, m.getErr
}

func (m *mockRepo) GetCommentsByTopicIDWithVotes(_ context.Context, _ int, _ *string) ([]comment.Comment, error) {
	return m.getResultWV, m.getErrWV
}

func TestGetCommentsByTopicResolver_Success(t *testing.T) {
	expected := []comment.Comment{
		{ID: 1, TopicID: 1, UserID: "u1", Content: "a"},
		{ID: 2, TopicID: 1, UserID: "u2", Content: "b"},
	}
	repo := &mockRepo{getResult: expected}
	r := NewGetCommentsByTopicResolver(repo)

	result, err := r.Resolve(context.Background(), GetCommentsByTopicQuery{TopicID: 1})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].Content != "a" || result[1].Content != "b" {
		t.Errorf("result = %+v, want Content=a, b", result)
	}
}

func TestGetCommentsByTopicResolver_Empty(t *testing.T) {
	repo := &mockRepo{getResult: []comment.Comment{}}
	r := NewGetCommentsByTopicResolver(repo)

	result, err := r.Resolve(context.Background(), GetCommentsByTopicQuery{TopicID: 1})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

func TestGetCommentsByTopicResolver_Error(t *testing.T) {
	repo := &mockRepo{getErr: errors.New("db down")}
	r := NewGetCommentsByTopicResolver(repo)

	_, err := r.Resolve(context.Background(), GetCommentsByTopicQuery{TopicID: 1})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetCommentsByTopicResolver_WithVotes(t *testing.T) {
	expected := []comment.Comment{
		{ID: 1, TopicID: 1, UserID: "u1", Content: "a", UpvoteCount: 3},
	}
	repo := &mockRepo{getResultWV: expected}
	r := NewGetCommentsByTopicResolver(repo)

	uid := "u2"
	result, err := r.Resolve(context.Background(), GetCommentsByTopicQuery{TopicID: 1, UserID: &uid})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(result) != 1 || result[0].UpvoteCount != 3 {
		t.Errorf("result = %+v, want UpvoteCount=3", result)
	}
}

func TestGetCommentsByTopicResolver_WithVotes_Error(t *testing.T) {
	repo := &mockRepo{getErrWV: errors.New("db down")}
	r := NewGetCommentsByTopicResolver(repo)

	uid := "u2"
	_, err := r.Resolve(context.Background(), GetCommentsByTopicQuery{TopicID: 1, UserID: &uid})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
