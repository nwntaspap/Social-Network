package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

type castGroupPostVoteStub struct {
	gotPostID   string
	gotReaction int
	castErr     error
	castCalled  bool
}

func (s *castGroupPostVoteStub) CreatePost(_ context.Context, _ *group.Post) error { return nil }
func (s *castGroupPostVoteStub) GetPostsByGroupID(_ context.Context, _ string, _ string, _, _ int) ([]group.Post, int, error) {
	return nil, 0, nil
}

func (s *castGroupPostVoteStub) GetPostVoteCounts(_ context.Context, _ string) (*group.VoteCounts, error) {
	return &group.VoteCounts{}, nil
}

func (s *castGroupPostVoteStub) CastPostVote(_ context.Context, _ string, postID string, reactionType int) error {
	s.castCalled = true
	s.gotPostID = postID
	s.gotReaction = reactionType
	return s.castErr
}

func TestCastGroupPostVoteHandler_Execute(t *testing.T) {
	ctx := context.Background()
	stub := &castGroupPostVoteStub{}
	h := NewCastGroupPostVoteHandler(stub)

	t.Run("valid like", func(t *testing.T) {
		stub.castCalled = false
		if err := h.Execute(ctx, CastGroupPostVoteCommand{UserID: "u1", PostID: "p1", ReactionType: 1}); err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if !stub.castCalled || stub.gotPostID != "p1" || stub.gotReaction != 1 {
			t.Errorf("stub = %+v, want called with p1/1", stub)
		}
	})

	t.Run("missing user", func(t *testing.T) {
		if err := h.Execute(ctx, CastGroupPostVoteCommand{PostID: "p1", ReactionType: 1}); !errors.Is(err, ErrUserIDRequired) {
			t.Errorf("err = %v, want ErrUserIDRequired", err)
		}
	})

	t.Run("missing post", func(t *testing.T) {
		if err := h.Execute(ctx, CastGroupPostVoteCommand{UserID: "u1", ReactionType: 1}); !errors.Is(err, group.ErrPostNotFound) {
			t.Errorf("err = %v, want ErrPostNotFound", err)
		}
	})

	t.Run("invalid reaction", func(t *testing.T) {
		if err := h.Execute(ctx, CastGroupPostVoteCommand{UserID: "u1", PostID: "p1", ReactionType: 5}); !errors.Is(err, group.ErrInvalidVoteValue) {
			t.Errorf("err = %v, want ErrInvalidVoteValue", err)
		}
	})
}

func (s *castGroupPostVoteStub) GetPostByID(_ context.Context, _ string) (*group.Post, error) {
	return nil, group.ErrPostNotFound
}
