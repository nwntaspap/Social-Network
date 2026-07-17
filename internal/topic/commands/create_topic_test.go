package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/topic"
)

type mockTopicRepo struct {
	createFn    func(ctx context.Context, t *topic.Topic, allowed []string) error
	updateFn    func(ctx context.Context, t *topic.Topic, allowed []string) error
	deleteFn    func(ctx context.Context, userID string, topicID int) error
	getByIDFn   func(ctx context.Context, id int, userID *string) (*topic.Topic, error)
	getImageFn  func(ctx context.Context, topicID int, userID string) (string, error)
	castVoteFn  func(ctx context.Context, userID string, topicID int, reaction int) error
	getCountsFn func(ctx context.Context, topicID int) (*topic.VoteCounts, error)
}

func (m *mockTopicRepo) CreateTopic(ctx context.Context, t *topic.Topic, allowed []string) error {
	if m.createFn != nil {
		return m.createFn(ctx, t, allowed)
	}
	t.ID = 1
	return nil
}

func (m *mockTopicRepo) UpdateTopic(_ context.Context, t *topic.Topic, _ []string) error {
	if m.updateFn != nil {
		return m.updateFn(context.TODO(), t, nil) //nolint:contextcheck // test mock, no parent ctx
	}
	return nil
}

func (m *mockTopicRepo) DeleteTopic(ctx context.Context, userID string, topicID int) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, topicID)
	}
	return nil
}

func (m *mockTopicRepo) GetTopicByID(ctx context.Context, id int, userID *string) (*topic.Topic, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id, userID)
	}
	return nil, topic.ErrTopicNotFound
}

func (m *mockTopicRepo) GetImagePathFromTopicID(ctx context.Context, topicID int, userID string) (string, error) {
	if m.getImageFn != nil {
		return m.getImageFn(ctx, topicID, userID)
	}
	return "", nil
}

func (m *mockTopicRepo) GetFeed(_ context.Context, _ string, _, _ int, _, _, _ string) ([]topic.Topic, int, error) {
	return nil, 0, nil
}

func (m *mockTopicRepo) GetTopicsByUserID(_ context.Context, _, _ string, _, _ int) ([]topic.Topic, int, error) {
	return nil, 0, nil
}

func (m *mockTopicRepo) GetTopicsByGroupID(_ context.Context, _ string, _, _ int) ([]topic.Topic, int, error) {
	return nil, 0, nil
}

func (m *mockTopicRepo) CastVote(ctx context.Context, userID string, topicID int, reaction int) error {
	if m.castVoteFn != nil {
		return m.castVoteFn(ctx, userID, topicID, reaction)
	}
	return nil
}
func (m *mockTopicRepo) DeleteVote(_ context.Context, _ string, _ int) error { return nil }
func (m *mockTopicRepo) GetVoteCounts(ctx context.Context, topicID int) (*topic.VoteCounts, error) {
	if m.getCountsFn != nil {
		return m.getCountsFn(ctx, topicID)
	}
	return &topic.VoteCounts{}, nil
}

type mockEventBus struct {
	eventType string
}

func (m *mockEventBus) Publish(_ context.Context, eventType string, _ any) error {
	m.eventType = eventType
	return nil
}

type mockImageStorage struct {
	uploadErr error
	deleteErr error
}

func (m *mockImageStorage) Upload(_ context.Context, _ []byte, _ string) error { return m.uploadErr }

func (m *mockImageStorage) Delete(_ context.Context, _ string) error { return m.deleteErr }

func TestCreateTopic_Success(t *testing.T) {
	bus := &mockEventBus{}
	h := NewCreateTopicHandler(&mockTopicRepo{}, bus, &mockImageStorage{})

	top, err := h.Execute(context.Background(), CreateTopicCommand{
		UserID:  "u1",
		Title:   "Hello",
		Content: "World",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if top.ID == 0 {
		t.Error("ID not set")
	}
	if bus.eventType != "post.created" {
		t.Errorf("event = %q, want %q", bus.eventType, "post.created")
	}
}

func TestCreateTopic_EmptyUser(t *testing.T) {
	h := NewCreateTopicHandler(&mockTopicRepo{}, &mockEventBus{}, &mockImageStorage{})

	_, err := h.Execute(context.Background(), CreateTopicCommand{Title: "X", Content: "Y"})
	if !errors.Is(err, topic.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

func TestCreateTopic_EmptyTitle(t *testing.T) {
	h := NewCreateTopicHandler(&mockTopicRepo{}, &mockEventBus{}, &mockImageStorage{})

	_, err := h.Execute(context.Background(), CreateTopicCommand{UserID: "u1", Content: "Y"})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestCreateTopic_RepoError(t *testing.T) {
	repo := &mockTopicRepo{
		createFn: func(_ context.Context, _ *topic.Topic, _ []string) error {
			return errors.New("db error")
		},
	}
	bus := &mockEventBus{}
	h := NewCreateTopicHandler(repo, bus, &mockImageStorage{})

	_, err := h.Execute(context.Background(), CreateTopicCommand{
		UserID: "u1", Title: "X", Content: "Y",
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
	if bus.eventType != "" {
		t.Errorf("event published after error: %q", bus.eventType)
	}
}
