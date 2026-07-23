package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
)

type mockVoteCountsRepo struct {
	result *comment.VoteCounts
	err    error
}

func (m *mockVoteCountsRepo) CreateComment(_ context.Context, _ *comment.Comment) error { return nil }

func (m *mockVoteCountsRepo) UpdateComment(_ context.Context, _ *comment.Comment) error { return nil }

func (m *mockVoteCountsRepo) DeleteComment(_ context.Context, _ string, _ int) error { return nil }

func (m *mockVoteCountsRepo) GetCommentByID(_ context.Context, _ int) (*comment.Comment, error) {
	return nil, comment.ErrCommentNotFound
}

func (m *mockVoteCountsRepo) GetCommentByIDWithVotes(_ context.Context, _ int, _ *string) (*comment.Comment, error) {
	return nil, comment.ErrCommentNotFound
}

func (m *mockVoteCountsRepo) GetCommentsByTopicID(_ context.Context, _ int) ([]comment.Comment, error) {
	return nil, nil
}

func (m *mockVoteCountsRepo) GetCommentsByTopicIDWithVotes(_ context.Context, _ int, _ *string) ([]comment.Comment, error) {
	return nil, nil
}

func (m *mockVoteCountsRepo) CastCommentVote(_ context.Context, _ string, _ int, _ int) error {
	return nil
}

func (m *mockVoteCountsRepo) GetVoteCounts(_ context.Context, _ int) (*comment.VoteCounts, error) {
	return m.result, m.err
}

func TestGetVoteCountsResolver_Success(t *testing.T) {
	expected := &comment.VoteCounts{Upvotes: 5, Downvotes: 2, Score: 3}
	repo := &mockVoteCountsRepo{result: expected}
	r := NewGetVoteCountsResolver(repo)

	result, err := r.Resolve(context.Background(), GetVoteCountsQuery{CommentID: 1})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.Upvotes != 5 || result.Downvotes != 2 || result.Score != 3 {
		t.Errorf("result = %+v, want Upvotes=5, Downvotes=2, Score=3", result)
	}
}

func TestGetVoteCountsResolver_ZeroVotes(t *testing.T) {
	expected := &comment.VoteCounts{Upvotes: 0, Downvotes: 0, Score: 0}
	repo := &mockVoteCountsRepo{result: expected}
	r := NewGetVoteCountsResolver(repo)

	result, err := r.Resolve(context.Background(), GetVoteCountsQuery{CommentID: 1})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.Upvotes != 0 || result.Downvotes != 0 || result.Score != 0 {
		t.Errorf("result = %+v, want all zeros", result)
	}
}

func TestGetVoteCountsResolver_Error(t *testing.T) {
	repo := &mockVoteCountsRepo{err: errors.New("db down")}
	r := NewGetVoteCountsResolver(repo)

	_, err := r.Resolve(context.Background(), GetVoteCountsQuery{CommentID: 1})
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
