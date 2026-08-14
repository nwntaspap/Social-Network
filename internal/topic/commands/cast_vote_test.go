package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/topic"
)

var topicWithAuthor = &mockTopicRepo{
	getByIDFn: func(_ context.Context, id int, _ *string) (*topic.Topic, error) {
		return &topic.Topic{ID: id, UserID: "author-1"}, nil
	},
}

func TestCastVote_Like(t *testing.T) {
	bus := &mockEventBus{}
	h := NewCastVoteHandler(topicWithAuthor, bus, &mockUserRepo{})

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID:       "u2",
		TopicID:      1,
		ReactionType: 1,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if bus.routingKey != "created" {
		t.Errorf("routingKey = %q, want %q", bus.routingKey, "created")
	}
}

func TestCastVote_Dislike(t *testing.T) {
	bus := &mockEventBus{}
	h := NewCastVoteHandler(topicWithAuthor, bus, &mockUserRepo{})

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID:       "u2",
		TopicID:      1,
		ReactionType: -1,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if bus.routingKey != "created" {
		t.Errorf("routingKey = %q, want %q", bus.routingKey, "created")
	}
}

func TestCastVote_EmptyUser(t *testing.T) {
	h := NewCastVoteHandler(&mockTopicRepo{}, &mockEventBus{}, &mockUserRepo{})

	err := h.Execute(context.Background(), CastVoteCommand{
		TopicID:      1,
		ReactionType: 1,
	})
	if !errors.Is(err, topic.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

func TestCastVote_ZeroTopicID(t *testing.T) {
	h := NewCastVoteHandler(&mockTopicRepo{}, &mockEventBus{}, &mockUserRepo{})

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID:       "u1",
		ReactionType: 1,
	})
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("err = %v, want ErrTopicNotFound", err)
	}
}

func TestCastVote_InvalidReaction(t *testing.T) {
	h := NewCastVoteHandler(&mockTopicRepo{}, &mockEventBus{}, &mockUserRepo{})

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID:       "u1",
		TopicID:      1,
		ReactionType: 0,
	})
	if !errors.Is(err, topic.ErrInvalidVoteValue) {
		t.Errorf("err = %v, want ErrInvalidVoteValue", err)
	}
}

func TestCastVote_RepoError(t *testing.T) {
	repo := &mockTopicRepo{
		castVoteFn: func(_ context.Context, _ string, _ int, _ int) (topic.VoteChange, error) {
			return topic.VoteChangeAdded, errors.New("db error")
		},
	}
	bus := &mockEventBus{}
	h := NewCastVoteHandler(repo, bus, &mockUserRepo{})

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID: "u1", TopicID: 1, ReactionType: 1,
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
	if bus.routingKey != "" {
		t.Errorf("event published after error: %q", bus.routingKey)
	}
}

func TestCastVote_ToggleOff(t *testing.T) {
	bus := &mockEventBus{}
	repo := &mockTopicRepo{
		castVoteFn: func(_ context.Context, _ string, _ int, _ int) (topic.VoteChange, error) {
			return topic.VoteChangeRemoved, nil
		},
	}
	h := NewCastVoteHandler(repo, bus, &mockUserRepo{})

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID: "u2", TopicID: 1, ReactionType: 1,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if bus.routingKey != "deleted" {
		t.Errorf("routingKey = %q, want %q", bus.routingKey, "deleted")
	}
	if bus.eventType != "post.vote.deleted" {
		t.Errorf("eventType = %q, want %q", bus.eventType, "post.vote.deleted")
	}
}
