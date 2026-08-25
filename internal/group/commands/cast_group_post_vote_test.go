package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type castGroupPostVoteStub struct {
	gotPostID   string
	gotReaction int
	castErr     error
	castCalled  bool
	post        *group.Post
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

func (s *castGroupPostVoteStub) GetPostByID(_ context.Context, _ string) (*group.Post, error) {
	if s.post != nil {
		return s.post, nil
	}
	return nil, group.ErrPostNotFound
}

type noopBus struct{}

func (noopBus) Publish(_ string, _ string, _ []byte) error { return nil }
func (noopBus) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}
func (noopBus) InitTopology(_ context.Context) error { return nil }

type noopUserRepo struct{}

func (noopUserRepo) Create(_ context.Context, _ *user.User) error { return nil }
func (noopUserRepo) GetByID(_ context.Context, _ string) (*user.User, error) {
	return &user.User{ID: "stub", Nickname: "stub"}, nil
}

func (noopUserRepo) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}

func (noopUserRepo) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}
func (noopUserRepo) Update(_ context.Context, _ *user.User) error            { return nil }
func (noopUserRepo) Delete(_ context.Context, _ string) error                { return nil }
func (noopUserRepo) TogglePrivacy(_ context.Context, _ string, _ bool) error { return nil }
func (noopUserRepo) ListAll(_ context.Context) ([]user.User, error)          { return nil, nil }

func TestCastGroupPostVoteHandler_Execute(t *testing.T) {
	ctx := context.Background()
	stub := &castGroupPostVoteStub{post: &group.Post{ID: "p1", AuthorID: "author1", Content: "hello"}}
	h := NewCastGroupPostVoteHandler(stub, noopBus{}, noopUserRepo{})

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
		stub2 := &castGroupPostVoteStub{}
		h2 := NewCastGroupPostVoteHandler(stub2, noopBus{}, noopUserRepo{})
		if err := h2.Execute(ctx, CastGroupPostVoteCommand{UserID: "u1", ReactionType: 1}); !errors.Is(err, group.ErrPostNotFound) {
			t.Errorf("err = %v, want ErrPostNotFound", err)
		}
	})

	t.Run("invalid reaction", func(t *testing.T) {
		if err := h.Execute(ctx, CastGroupPostVoteCommand{UserID: "u1", PostID: "p1", ReactionType: 5}); !errors.Is(err, group.ErrInvalidVoteValue) {
			t.Errorf("err = %v, want ErrInvalidVoteValue", err)
		}
	})
}
