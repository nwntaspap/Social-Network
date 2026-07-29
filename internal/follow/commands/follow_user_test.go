package commands

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"social-network/internal/follow"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

var errUserNotFound = errors.New("user not found")

type mockUserRepo struct{}

func (m *mockUserRepo) Create(_ context.Context, _ *user.User) error { return nil }
func (m *mockUserRepo) GetByID(_ context.Context, id string) (*user.User, error) {
	return &user.User{Nickname: id + "-name", AvatarPath: ""}, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, errUserNotFound
}

func (m *mockUserRepo) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, errUserNotFound
}
func (m *mockUserRepo) Update(_ context.Context, _ *user.User) error            { return nil }
func (m *mockUserRepo) TogglePrivacy(_ context.Context, _ string, _ bool) error { return nil }
func (m *mockUserRepo) ListAll(_ context.Context) ([]user.User, error)          { return nil, nil }

type mockRepo struct {
	createFollowErr        error
	createFollowRequestErr error
}

func (m *mockRepo) CreateFollow(_ context.Context, _ *follow.Follow) error {
	return m.createFollowErr
}
func (m *mockRepo) DeleteFollow(_ context.Context, _, _ string) error { return nil }
func (m *mockRepo) GetFollowers(_ context.Context, _ string) ([]follow.Follow, error) {
	return nil, nil
}

func (m *mockRepo) GetFollowing(_ context.Context, _ string) ([]follow.Follow, error) {
	return nil, nil
}

func (m *mockRepo) CreateFollowRequest(_ context.Context, _ *follow.Request) error {
	return m.createFollowRequestErr
}

func (m *mockRepo) DeleteFollowRequest(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockRepo) GetPendingRequests(_ context.Context, _ string) ([]follow.Request, error) {
	return nil, nil
}

func (m *mockRepo) AreConnected(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (m *mockRepo) GetFollowerCount(_ context.Context, _ string) (int, error)  { return 0, nil }
func (m *mockRepo) GetFollowingCount(_ context.Context, _ string) (int, error) { return 0, nil }

type mockPrivacy struct {
	isPrivate bool
	err       error
}

func (m *mockPrivacy) IsPrivate(_ context.Context, _ string) (bool, error) {
	return m.isPrivate, m.err
}

type mockBus struct {
	routingKey string
	body       []byte
	publishErr error
}

func (m *mockBus) Publish(exchange, routingKey string, body []byte) error {
	m.routingKey = routingKey
	m.body = body
	return m.publishErr
}

func (m *mockBus) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}

func (m *mockBus) InitTopology(_ context.Context) error {
	return nil
}

func TestFollowUserHandler_PublicUser(t *testing.T) {
	repo := &mockRepo{}
	privacy := &mockPrivacy{isPrivate: false}
	bus := &mockBus{}
	h := NewFollowUserHandler(repo, privacy, bus, &mockUserRepo{})

	result, err := h.Execute(context.Background(), FollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-2",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result != FollowedDirect {
		t.Errorf("result = %q, want %q", result, FollowedDirect)
	}
	if bus.routingKey != "follow.accepted" {
		t.Errorf("routingKey = %q, want %q", bus.routingKey, "follow.accepted")
	}
	var env eventbus.Envelope
	if err := json.Unmarshal(bus.body, &env); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if env.RecipientID != "user-2" || env.ActorID != "user-1" {
		t.Errorf("payload = %+v, want RecipientID=user-2 ActorID=user-1", env)
	}
}

func TestFollowUserHandler_PrivateUser(t *testing.T) {
	repo := &mockRepo{}
	privacy := &mockPrivacy{isPrivate: true}
	bus := &mockBus{}
	h := NewFollowUserHandler(repo, privacy, bus, &mockUserRepo{})

	result, err := h.Execute(context.Background(), FollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-2",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result != FollowPending {
		t.Errorf("result = %q, want %q", result, FollowPending)
	}
	if bus.routingKey != "follow.requested" {
		t.Errorf("routingKey = %q, want %q", bus.routingKey, "follow.requested")
	}
	var env eventbus.Envelope
	if err := json.Unmarshal(bus.body, &env); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if env.RecipientID != "user-2" || env.ActorID != "user-1" {
		t.Errorf("payload = %+v, want RecipientID=user-2 ActorID=user-1", env)
	}
}

func TestFollowUserHandler_SelfFollow(t *testing.T) {
	h := NewFollowUserHandler(&mockRepo{}, &mockPrivacy{}, &mockBus{}, &mockUserRepo{})

	_, err := h.Execute(context.Background(), FollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-1",
	})
	if !errors.Is(err, ErrSelfFollow) {
		t.Errorf("Execute() error = %v, want %v", err, ErrSelfFollow)
	}
}

func TestFollowUserHandler_PrivacyCheckError(t *testing.T) {
	privacy := &mockPrivacy{err: errors.New("db down")}
	h := NewFollowUserHandler(&mockRepo{}, privacy, &mockBus{}, &mockUserRepo{})

	_, err := h.Execute(context.Background(), FollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-2",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}

func TestFollowUserHandler_RepoCreateFollowError(t *testing.T) {
	repo := &mockRepo{createFollowErr: errors.New("insert failed")}
	privacy := &mockPrivacy{isPrivate: false}
	bus := &mockBus{}
	h := NewFollowUserHandler(repo, privacy, bus, &mockUserRepo{})

	_, err := h.Execute(context.Background(), FollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-2",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if bus.routingKey != "" {
		t.Errorf("event published after repo error: %q", bus.routingKey)
	}
}
