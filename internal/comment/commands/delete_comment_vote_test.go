package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
)

func TestDeleteCommentVote_Success(t *testing.T) {
	repo := &mockDeleteVoteRepo{}
	bus := &mockBus{}
	h := NewDeleteCommentVoteHandler(repo, bus)

	err := h.Execute(context.Background(), DeleteCommentVoteCommand{
		UserID:    "u1",
		CommentID: 1,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if bus.eventType != "comment.vote.deleted" {
		t.Errorf("event = %q, want %q", bus.eventType, "comment.vote.deleted")
	}
}

func TestDeleteCommentVote_EmptyUserID(t *testing.T) {
	h := NewDeleteCommentVoteHandler(&mockDeleteVoteRepo{}, &mockBus{})

	err := h.Execute(context.Background(), DeleteCommentVoteCommand{
		CommentID: 1,
	})
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("error = %v, want ErrEmptyUserID", err)
	}
}

func TestDeleteCommentVote_ZeroCommentID(t *testing.T) {
	h := NewDeleteCommentVoteHandler(&mockDeleteVoteRepo{}, &mockBus{})

	err := h.Execute(context.Background(), DeleteCommentVoteCommand{
		UserID: "u1",
	})
	if !errors.Is(err, ErrEmptyCommentID) {
		t.Errorf("error = %v, want ErrEmptyCommentID", err)
	}
}

func TestDeleteCommentVote_RepoError(t *testing.T) {
	repo := &mockDeleteVoteRepo{deleteVoteErr: errors.New("db down")}
	bus := &mockBus{}
	h := NewDeleteCommentVoteHandler(repo, bus)

	err := h.Execute(context.Background(), DeleteCommentVoteCommand{
		UserID: "u1", CommentID: 1,
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if bus.eventType != "" {
		t.Errorf("event published after error: %q", bus.eventType)
	}
}

func TestDeleteCommentVote_VoteNotFound(t *testing.T) {
	repo := &mockDeleteVoteRepo{deleteVoteErr: comment.ErrVoteNotFound}
	h := NewDeleteCommentVoteHandler(repo, &mockBus{})

	err := h.Execute(context.Background(), DeleteCommentVoteCommand{
		UserID: "u1", CommentID: 1,
	})
	if !errors.Is(err, comment.ErrVoteNotFound) {
		t.Errorf("error = %v, want ErrVoteNotFound", err)
	}
}

type mockDeleteVoteRepo struct {
	deleteVoteErr error
}

func (m *mockDeleteVoteRepo) CreateComment(_ context.Context, _ *comment.Comment) error { return nil }

func (m *mockDeleteVoteRepo) UpdateComment(_ context.Context, _ *comment.Comment) error { return nil }

func (m *mockDeleteVoteRepo) DeleteComment(_ context.Context, _ string, _ int) error { return nil }

func (m *mockDeleteVoteRepo) GetCommentByID(_ context.Context, _ int) (*comment.Comment, error) {
	return nil, comment.ErrCommentNotFound
}

func (m *mockDeleteVoteRepo) GetCommentByIDWithVotes(_ context.Context, _ int, _ *string) (*comment.Comment, error) {
	return nil, comment.ErrCommentNotFound
}

func (m *mockDeleteVoteRepo) GetCommentsByTopicID(_ context.Context, _ int) ([]comment.Comment, error) {
	return nil, nil
}

func (m *mockDeleteVoteRepo) GetCommentsByTopicIDWithVotes(_ context.Context, _ int, _ *string) ([]comment.Comment, error) {
	return nil, nil
}

func (m *mockDeleteVoteRepo) CastCommentVote(_ context.Context, _ string, _ int, _ int) error {
	return nil
}

func (m *mockDeleteVoteRepo) DeleteCommentVote(_ context.Context, _ string, _ int) error {
	return m.deleteVoteErr
}

func (m *mockDeleteVoteRepo) GetVoteCounts(_ context.Context, _ int) (*comment.VoteCounts, error) {
	return &comment.VoteCounts{}, nil
}

func (m *mockDeleteVoteRepo) GetCommentCount(_ context.Context, _ string) (int, error) {
	return 0, nil
}
