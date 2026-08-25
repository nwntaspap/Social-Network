package commands

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"social-network/internal/platform/eventbus"
	"social-network/internal/topic"
	"social-network/internal/user"
)

type mockUserRepo struct{}

func (m *mockUserRepo) Create(_ context.Context, _ *user.User) error { return nil }
func (m *mockUserRepo) GetByID(_ context.Context, id string) (*user.User, error) {
	return &user.User{Nickname: id + "-name", AvatarPath: ""}, nil
}

var errUserNotFound = errors.New("user not found")

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, errUserNotFound
}

func (m *mockUserRepo) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, errUserNotFound
}
func (m *mockUserRepo) Update(_ context.Context, _ *user.User) error            { return nil }
func (m *mockUserRepo) TogglePrivacy(_ context.Context, _ string, _ bool) error { return nil }
func (m *mockUserRepo) ListAll(_ context.Context) ([]user.User, error)          { return nil, nil }

type mockTopicRepo struct {
	createFn     func(ctx context.Context, t *topic.Topic, allowed []string) error
	updateFn     func(ctx context.Context, t *topic.Topic, allowed []string) error
	deleteFn     func(ctx context.Context, userID string, topicID int) error
	getByIDFn    func(ctx context.Context, id int, userID *string) (*topic.Topic, error)
	getImageFn   func(ctx context.Context, topicID int, userID string) (string, error)
	castVoteFn   func(ctx context.Context, userID string, topicID int, reaction int) (topic.VoteChange, error)
	deleteVoteFn func(ctx context.Context, userID string, topicID int) error
	getCountsFn  func(ctx context.Context, topicID int) (*topic.VoteCounts, error)
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

func (m *mockTopicRepo) CastVote(ctx context.Context, userID string, topicID int, reaction int) (topic.VoteChange, error) {
	if m.castVoteFn != nil {
		return m.castVoteFn(ctx, userID, topicID, reaction)
	}
	return topic.VoteChangeAdded, nil
}

func (m *mockTopicRepo) DeleteVote(ctx context.Context, userID string, topicID int) error {
	if m.deleteVoteFn != nil {
		return m.deleteVoteFn(ctx, userID, topicID)
	}
	return nil
}

func (m *mockTopicRepo) GetVoteCounts(ctx context.Context, topicID int) (*topic.VoteCounts, error) {
	if m.getCountsFn != nil {
		return m.getCountsFn(ctx, topicID)
	}
	return &topic.VoteCounts{}, nil
}
func (m *mockTopicRepo) GetPostCount(_ context.Context, _ string) (int, error) { return 0, nil }
func (m *mockTopicRepo) GetVoteCount(_ context.Context, _ string) (int, error) { return 0, nil }

type mockEventBus struct {
	routingKey string
	eventType  string
	calls      []mockPublishCall
}

type mockPublishCall struct {
	routingKey string
	eventType  string
}

func (m *mockEventBus) Publish(exchange string, routingKey string, body []byte) error {
	m.routingKey = routingKey
	var n eventbus.Notification
	if err := json.Unmarshal(body, &n); err == nil {
		m.eventType = n.Type
	}
	m.calls = append(m.calls, mockPublishCall{routingKey: routingKey, eventType: n.Type})
	return nil
}

func (m *mockEventBus) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}

func (m *mockEventBus) InitTopology(_ context.Context) error {
	return nil
}

type mockImageStorage struct {
	uploadErr error
	deleteErr error
}

func (m *mockImageStorage) Upload(_ context.Context, _ []byte, _ string) error { return m.uploadErr }

func (m *mockImageStorage) Delete(_ context.Context, _ string) error { return m.deleteErr }

func TestCreateTopic_Success(t *testing.T) {
	h := NewCreateTopicHandler(&mockTopicRepo{}, &mockImageStorage{})

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
}

func TestCreateTopic_EmptyUser(t *testing.T) {
	h := NewCreateTopicHandler(&mockTopicRepo{}, &mockImageStorage{})

	_, err := h.Execute(context.Background(), CreateTopicCommand{Title: "X", Content: "Y"})
	if !errors.Is(err, topic.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

func TestCreateTopic_EmptyTitle(t *testing.T) {
	h := NewCreateTopicHandler(&mockTopicRepo{}, &mockImageStorage{})

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
	h := NewCreateTopicHandler(repo, &mockImageStorage{})

	_, err := h.Execute(context.Background(), CreateTopicCommand{
		UserID: "u1", Title: "X", Content: "Y",
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestCreateTopic_RejectsInvalidImageHeader(t *testing.T) {
	h := NewCreateTopicHandler(&mockTopicRepo{}, &mockImageStorage{})

	_, err := h.Execute(context.Background(), CreateTopicCommand{
		UserID:        "u1",
		Title:         "X",
		Content:       "Y",
		ImageData:     []byte{0x00, 0x01, 0x02, 0x03},
		ImageFileName: "evil.png",
	})
	if err == nil {
		t.Fatal("expected error for invalid image header")
	}
}
