package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/topic"
)

func TestCastVote_Like(t *testing.T) {
	bus := &mockEventBus{}
	h := NewCastVoteHandler(&mockTopicRepo{}, bus)

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID:       "u2",
		TopicID:      1,
		ReactionType: 1,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if bus.eventType != "post.liked" {
		t.Errorf("event = %q, want %q", bus.eventType, "post.liked")
	}
}

func TestCastVote_Dislike(t *testing.T) {
	bus := &mockEventBus{}
	h := NewCastVoteHandler(&mockTopicRepo{}, bus)

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID:       "u2",
		TopicID:      1,
		ReactionType: -1,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if bus.eventType != "post.unliked" {
		t.Errorf("event = %q, want %q", bus.eventType, "post.unliked")
	}
}

func TestCastVote_EmptyUser(t *testing.T) {
	h := NewCastVoteHandler(&mockTopicRepo{}, &mockEventBus{})

	err := h.Execute(context.Background(), CastVoteCommand{
		TopicID:      1,
		ReactionType: 1,
	})
	if !errors.Is(err, topic.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

func TestCastVote_ZeroTopicID(t *testing.T) {
	h := NewCastVoteHandler(&mockTopicRepo{}, &mockEventBus{})

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID:       "u1",
		ReactionType: 1,
	})
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("err = %v, want ErrTopicNotFound", err)
	}
}

func TestCastVote_InvalidReaction(t *testing.T) {
	h := NewCastVoteHandler(&mockTopicRepo{}, &mockEventBus{})

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
		castVoteFn: func(_ context.Context, _ string, _ int, _ int) error {
			return errors.New("db error")
		},
	}
	bus := &mockEventBus{}
	h := NewCastVoteHandler(repo, bus)

	err := h.Execute(context.Background(), CastVoteCommand{
		UserID: "u1", TopicID: 1, ReactionType: 1,
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
	if bus.eventType != "" {
		t.Errorf("event published after error: %q", bus.eventType)
	}
}
