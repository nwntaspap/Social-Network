package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/comment"
	"social-network/internal/user"
)

type commentUserRepo struct{}

func (m *commentUserRepo) Create(_ context.Context, _ *user.User) error { return nil }
func (m *commentUserRepo) GetByID(_ context.Context, id string) (*user.User, error) {
	return &user.User{Nickname: id + "-name", AvatarPath: ""}, nil
}

var errUserNotFound = errors.New("user not found")

func (m *commentUserRepo) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, errUserNotFound
}

func (m *commentUserRepo) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, errUserNotFound
}
func (m *commentUserRepo) Update(_ context.Context, _ *user.User) error            { return nil }
func (m *commentUserRepo) TogglePrivacy(_ context.Context, _ string, _ bool) error { return nil }
func (m *commentUserRepo) ListAll(_ context.Context) ([]user.User, error)          { return nil, nil }

type mockVoteRepo struct {
	castVoteErr     error
	getCountsResult *comment.VoteCounts
	getCountsErr    error
}

func (m *mockVoteRepo) CreateComment(_ context.Context, _ *comment.Comment) error { return nil }
func (m *mockVoteRepo) UpdateComment(_ context.Context, _ *comment.Comment) error { return nil }
func (m *mockVoteRepo) DeleteComment(_ context.Context, _ string, _ int) error    { return nil }
func (m *mockVoteRepo) GetCommentByID(_ context.Context, _ int) (*comment.Comment, error) {
	return &comment.Comment{ID: 1, UserID: "comment-author", Content: "hello"}, nil
}

func (m *mockVoteRepo) GetCommentByIDWithVotes(_ context.Context, _ int, _ *string) (*comment.Comment, error) {
	return nil, comment.ErrCommentNotFound
}

func (m *mockVoteRepo) GetCommentsByTopicID(_ context.Context, _ int) ([]comment.Comment, error) {
	return []comment.Comment{}, nil
}

func (m *mockVoteRepo) GetCommentsByTopicIDWithVotes(_ context.Context, _ int, _ *string) ([]comment.Comment, error) {
	return []comment.Comment{}, nil
}

func (m *mockVoteRepo) CastCommentVote(_ context.Context, _ string, _ int, _ int) error {
	return m.castVoteErr
}

func (m *mockVoteRepo) DeleteCommentVote(_ context.Context, _ string, _ int) error {
	return nil
}

func (m *mockVoteRepo) GetVoteCounts(_ context.Context, _ int) (*comment.VoteCounts, error) {
	return m.getCountsResult, m.getCountsErr
}

func (m *mockVoteRepo) GetCommentCount(_ context.Context, _ string) (int, error) { return 0, nil }

func TestCastCommentVote_Success(t *testing.T) {
	repo := &mockVoteRepo{}
	h := NewCastCommentVoteHandler(repo, &mockBus{}, &commentUserRepo{})
	err := h.Execute(context.Background(), CastCommentVoteCommand{
		UserID:       "u1",
		CommentID:    1,
		ReactionType: 1,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestCastCommentVote_Downvote(t *testing.T) {
	repo := &mockVoteRepo{}
	h := NewCastCommentVoteHandler(repo, &mockBus{}, &commentUserRepo{})
	err := h.Execute(context.Background(), CastCommentVoteCommand{
		UserID:       "u1",
		CommentID:    1,
		ReactionType: -1,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestCastCommentVote_EmptyUserID(t *testing.T) {
	repo := &mockVoteRepo{}
	h := NewCastCommentVoteHandler(repo, &mockBus{}, &commentUserRepo{})
	err := h.Execute(context.Background(), CastCommentVoteCommand{
		CommentID:    1,
		ReactionType: 1,
	})
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("error = %v, want ErrEmptyUserID", err)
	}
}

func TestCastCommentVote_ZeroCommentID(t *testing.T) {
	repo := &mockVoteRepo{}
	h := NewCastCommentVoteHandler(repo, &mockBus{}, &commentUserRepo{})
	err := h.Execute(context.Background(), CastCommentVoteCommand{
		UserID:       "u1",
		ReactionType: 1,
	})
	if !errors.Is(err, ErrEmptyCommentID) {
		t.Errorf("error = %v, want ErrEmptyCommentID", err)
	}
}

func TestCastCommentVote_InvalidReactionType(t *testing.T) {
	repo := &mockVoteRepo{}
	h := NewCastCommentVoteHandler(repo, &mockBus{}, &commentUserRepo{})
	err := h.Execute(context.Background(), CastCommentVoteCommand{
		UserID:       "u1",
		CommentID:    1,
		ReactionType: 0,
	})
	if !errors.Is(err, ErrInvalidReactionType) {
		t.Errorf("error = %v, want ErrInvalidReactionType", err)
	}
}

func TestCastCommentVote_InvalidReactionType2(t *testing.T) {
	repo := &mockVoteRepo{}
	h := NewCastCommentVoteHandler(repo, &mockBus{}, &commentUserRepo{})
	err := h.Execute(context.Background(), CastCommentVoteCommand{
		UserID:       "u1",
		CommentID:    1,
		ReactionType: 5,
	})
	if !errors.Is(err, ErrInvalidReactionType) {
		t.Errorf("error = %v, want ErrInvalidReactionType", err)
	}
}

func TestCastCommentVote_RepoError(t *testing.T) {
	repo := &mockVoteRepo{castVoteErr: errors.New("db down")}
	h := NewCastCommentVoteHandler(repo, &mockBus{}, &commentUserRepo{})
	err := h.Execute(context.Background(), CastCommentVoteCommand{
		UserID:       "u1",
		CommentID:    1,
		ReactionType: 1,
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}
