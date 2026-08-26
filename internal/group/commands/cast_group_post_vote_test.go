package commands

import (
	"context"
	"encoding/json"
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
	voteChange  group.VoteChange
}

func (s *castGroupPostVoteStub) CreatePost(_ context.Context, _ *group.Post) error { return nil }
func (s *castGroupPostVoteStub) GetPostsByGroupID(_ context.Context, _ string, _ string, _, _ int) ([]group.Post, int, error) {
	return nil, 0, nil
}

func (s *castGroupPostVoteStub) GetPostVoteCounts(_ context.Context, _ string) (*group.VoteCounts, error) {
	return &group.VoteCounts{}, nil
}

func (s *castGroupPostVoteStub) CastPostVote(_ context.Context, _ string, postID string, reactionType int) (group.VoteChange, error) {
	s.castCalled = true
	s.gotPostID = postID
	s.gotReaction = reactionType
	return s.voteChange, s.castErr
}

func (s *castGroupPostVoteStub) GetPostByID(_ context.Context, _ string) (*group.Post, error) {
	if s.post != nil {
		return s.post, nil
	}
	return nil, group.ErrPostNotFound
}

type spyBus struct {
	calls []spyCall
}

type spyCall struct {
	routingKey string
	eventType  string
}

func (b *spyBus) Publish(_ string, routingKey string, body []byte) error {
	var n eventbus.Notification
	if err := json.Unmarshal(body, &n); err == nil {
		b.calls = append(b.calls, spyCall{routingKey: routingKey, eventType: n.Type})
	}
	return nil
}

func (b *spyBus) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}
func (b *spyBus) InitTopology(_ context.Context) error { return nil }

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
	h := NewCastGroupPostVoteHandler(stub, &spyBus{}, noopUserRepo{})

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
		h2 := NewCastGroupPostVoteHandler(stub2, &spyBus{}, noopUserRepo{})
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

func TestCastGroupPostVote_ChangeVote(t *testing.T) {
	bus := &spyBus{}
	stub := &castGroupPostVoteStub{
		post:       &group.Post{ID: "p1", AuthorID: "author1", Content: "hello"},
		voteChange: group.VoteChangeAdded,
	}
	h := NewCastGroupPostVoteHandler(stub, bus, noopUserRepo{})

	err := h.Execute(context.Background(), CastGroupPostVoteCommand{
		UserID: "u1", PostID: "p1", ReactionType: -1,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(bus.calls) < 2 {
		t.Fatalf("expected at least 2 publish calls, got %d", len(bus.calls))
	}
	if bus.calls[0].routingKey != "deleted" {
		t.Errorf("first call routingKey = %q, want %q", bus.calls[0].routingKey, "deleted")
	}
	if bus.calls[0].eventType != "post.vote.deleted" {
		t.Errorf("first call eventType = %q, want %q", bus.calls[0].eventType, "post.vote.deleted")
	}
	if bus.calls[1].routingKey != "created" {
		t.Errorf("second call routingKey = %q, want %q", bus.calls[1].routingKey, "created")
	}
	if bus.calls[1].eventType != "post.disliked" {
		t.Errorf("second call eventType = %q, want %q", bus.calls[1].eventType, "post.disliked")
	}
}

func TestCastGroupPostVote_ToggleOff(t *testing.T) {
	bus := &spyBus{}
	stub := &castGroupPostVoteStub{
		post:       &group.Post{ID: "p1", AuthorID: "author1", Content: "hello"},
		voteChange: group.VoteChangeRemoved,
	}
	h := NewCastGroupPostVoteHandler(stub, bus, noopUserRepo{})

	err := h.Execute(context.Background(), CastGroupPostVoteCommand{
		UserID: "u1", PostID: "p1", ReactionType: 1,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(bus.calls) != 1 {
		t.Fatalf("expected 1 publish call, got %d", len(bus.calls))
	}
	if bus.calls[0].routingKey != "deleted" {
		t.Errorf("routingKey = %q, want %q", bus.calls[0].routingKey, "deleted")
	}
	if bus.calls[0].eventType != "post.vote.deleted" {
		t.Errorf("eventType = %q, want %q", bus.calls[0].eventType, "post.vote.deleted")
	}
}
