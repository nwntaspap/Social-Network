package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/topic"
)

func TestDeleteVote_Success(t *testing.T) {
	bus := &mockEventBus{}
	h := NewDeleteVoteHandler(&mockTopicRepo{}, bus)

	err := h.Execute(context.Background(), DeleteVoteCommand{
		UserID:  "u2",
		TopicID: 1,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if bus.eventType != "post.unliked" {
		t.Errorf("event = %q, want %q", bus.eventType, "post.unliked")
	}
}

func TestDeleteVote_EmptyUser(t *testing.T) {
	h := NewDeleteVoteHandler(&mockTopicRepo{}, &mockEventBus{})

	err := h.Execute(context.Background(), DeleteVoteCommand{
		TopicID: 1,
	})
	if !errors.Is(err, topic.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

func TestDeleteVote_ZeroTopicID(t *testing.T) {
	h := NewDeleteVoteHandler(&mockTopicRepo{}, &mockEventBus{})

	err := h.Execute(context.Background(), DeleteVoteCommand{
		UserID: "u1",
	})
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("err = %v, want ErrTopicNotFound", err)
	}
}

func TestDeleteVote_RepoError(t *testing.T) {
	repo := &mockTopicRepo{
		deleteVoteFn: func(_ context.Context, _ string, _ int) error {
			return errors.New("db error")
		},
	}
	bus := &mockEventBus{}
	h := NewDeleteVoteHandler(repo, bus)

	err := h.Execute(context.Background(), DeleteVoteCommand{
		UserID: "u1", TopicID: 1,
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
	if bus.eventType != "" {
		t.Errorf("event published after error: %q", bus.eventType)
	}
}
