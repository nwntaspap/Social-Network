package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/topic"
)

func TestUpdateTopic_Success(t *testing.T) {
	repo := &mockTopicRepo{
		getByIDFn: func(_ context.Context, id int, _ *string) (*topic.Topic, error) {
			return &topic.Topic{ID: id, UserID: "u1"}, nil
		},
	}
	h := NewUpdateTopicHandler(repo, &mockImageStorage{})

	top, err := h.Execute(context.Background(), UpdateTopicCommand{
		TopicID:    1,
		UserID:     "u1",
		Title:      "Updated",
		Content:    "New content",
		Visibility: topic.VisibilityFollowers,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if top.Title != "Updated" {
		t.Errorf("Title = %q, want %q", top.Title, "Updated")
	}
}

func TestUpdateTopic_EmptyUser(t *testing.T) {
	h := NewUpdateTopicHandler(&mockTopicRepo{}, &mockImageStorage{})

	_, err := h.Execute(context.Background(), UpdateTopicCommand{TopicID: 1, Title: "X"})
	if !errors.Is(err, topic.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

func TestUpdateTopic_ZeroID(t *testing.T) {
	h := NewUpdateTopicHandler(&mockTopicRepo{}, &mockImageStorage{})

	_, err := h.Execute(context.Background(), UpdateTopicCommand{UserID: "u1", Title: "X"})
	if !errors.Is(err, topic.ErrTopicNotFound) {
		t.Errorf("err = %v, want ErrTopicNotFound", err)
	}
}
