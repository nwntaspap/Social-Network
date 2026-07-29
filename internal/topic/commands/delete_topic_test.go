package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/topic"
)

func TestDeleteTopic_Success(t *testing.T) {
	bus := &mockEventBus{}
	img := &mockImageStorage{}
	h := NewDeleteTopicHandler(&mockTopicRepo{}, bus, img)

	err := h.Execute(context.Background(), DeleteTopicCommand{
		TopicID: 1,
		UserID:  "u1",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if bus.routingKey != "post.deleted" {
		t.Errorf("routingKey = %q, want %q", bus.routingKey, "post.deleted")
	}
}

func TestDeleteTopic_EmptyUser(t *testing.T) {
	h := NewDeleteTopicHandler(&mockTopicRepo{}, &mockEventBus{}, &mockImageStorage{})

	err := h.Execute(context.Background(), DeleteTopicCommand{TopicID: 1})
	if !errors.Is(err, topic.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

func TestDeleteTopic_ZeroID(t *testing.T) {
	h := NewDeleteTopicHandler(&mockTopicRepo{}, &mockEventBus{}, &mockImageStorage{})

	err := h.Execute(context.Background(), DeleteTopicCommand{UserID: "u1"})
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("err = %v, want ErrTopicNotFound", err)
	}
}

func TestDeleteTopic_RepoError(t *testing.T) {
	repo := &mockTopicRepo{}
	repo.deleteFn = func(_ context.Context, _ string, _ int) error {
		return errors.New("db error")
	}
	bus := &mockEventBus{}
	h := NewDeleteTopicHandler(repo, bus, &mockImageStorage{})

	err := h.Execute(context.Background(), DeleteTopicCommand{
		TopicID: 1, UserID: "u1",
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
	if bus.routingKey != "" {
		t.Errorf("event published after error: %q", bus.routingKey)
	}
}
