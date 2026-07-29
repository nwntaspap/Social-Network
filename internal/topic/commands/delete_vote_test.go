package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/topic"
)

var topicWithAuthor2 = &mockTopicRepo{
	getByIDFn: func(_ context.Context, id int, _ *string) (*topic.Topic, error) {
		return &topic.Topic{ID: id, UserID: "author-1"}, nil
	},
}

func TestDeleteVote_Success(t *testing.T) {
	bus := &mockEventBus{}
	h := NewDeleteVoteHandler(topicWithAuthor2, bus)

	err := h.Execute(context.Background(), DeleteVoteCommand{
		UserID:  "u2",
		TopicID: 1,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if bus.routingKey != "post.liked.deleted" {
		t.Errorf("routingKey = %q, want %q", bus.routingKey, "post.liked.deleted")
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
	if bus.routingKey != "" {
		t.Errorf("event published after error: %q", bus.routingKey)
	}
}
