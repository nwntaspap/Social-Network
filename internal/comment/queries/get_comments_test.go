package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
	"social-network/internal/topic"
)

type fakeTopicChecker struct {
	err error
}

func (f *fakeTopicChecker) GetTopicByID(_ context.Context, _ int, _ *string) (*topic.Topic, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &topic.Topic{}, nil
}

type mockRepo struct {
	getResult     []comment.Comment
	getErr        error
	getByIDResult *comment.Comment
	getByIDErr    error

	getByIDWithVotesResult *comment.Comment
	getByIDWithVotesErr    error
	getResultWV            []comment.Comment
	getErrWV               error
}

func (m *mockRepo) CreateComment(_ context.Context, _ *comment.Comment) error { return nil }
func (m *mockRepo) UpdateComment(_ context.Context, _ *comment.Comment) error { return nil }
func (m *mockRepo) DeleteComment(_ context.Context, _ string, _ int) error    { return nil }

func (m *mockRepo) GetCommentByID(_ context.Context, _ int) (*comment.Comment, error) {
	return m.getByIDResult, m.getByIDErr
}

func (m *mockRepo) GetCommentByIDWithVotes(_ context.Context, _ int, _ *string) (*comment.Comment, error) {
	return m.getByIDWithVotesResult, m.getByIDWithVotesErr
}

func (m *mockRepo) GetCommentsByTopicID(_ context.Context, _ int) ([]comment.Comment, error) {
	return m.getResult, m.getErr
}

func (m *mockRepo) GetCommentsByTopicIDWithVotes(_ context.Context, _ int, _ *string) ([]comment.Comment, error) {
	return m.getResultWV, m.getErrWV
}

func (m *mockRepo) CastCommentVote(_ context.Context, _ string, _ int, _ int) (comment.VoteChange, error) {
	return comment.VoteChangeAdded, nil
}

func (m *mockRepo) DeleteCommentVote(_ context.Context, _ string, _ int) error {
	return nil
}

func (m *mockRepo) GetVoteCounts(_ context.Context, _ int) (*comment.VoteCounts, error) {
	return &comment.VoteCounts{}, nil
}

func (m *mockRepo) GetCommentCount(_ context.Context, _ string) (int, error) { return 0, nil }

func TestGetCommentsByTopicResolver_Success(t *testing.T) {
	expected := []comment.Comment{
		{ID: 1, TopicID: 1, UserID: "u1", Content: "a"},
		{ID: 2, TopicID: 1, UserID: "u2", Content: "b"},
	}
	repo := &mockRepo{getResult: expected}
	r := NewGetCommentsByTopicResolver(repo, &fakeTopicChecker{})

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
	r := NewGetCommentsByTopicResolver(repo, &fakeTopicChecker{})

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
	r := NewGetCommentsByTopicResolver(repo, &fakeTopicChecker{})

	_, err := r.Resolve(context.Background(), GetCommentsByTopicQuery{TopicID: 1})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}

func TestGetCommentsByTopicResolver_HiddenTopicDenied(t *testing.T) {
	repo := &mockRepo{getResult: []comment.Comment{{ID: 1, TopicID: 1}}}
	r := NewGetCommentsByTopicResolver(repo, &fakeTopicChecker{err: topic.ErrTopicNotFound})

	_, err := r.Resolve(context.Background(), GetCommentsByTopicQuery{TopicID: 1})
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("Resolve() error = %v, want ErrTopicNotFound", err)
	}
}
