package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/follow"
)

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

type mockPrivacy struct {
	isPrivate bool
	err       error
}

func (m *mockPrivacy) IsPrivate(_ context.Context, _ string) (bool, error) {
	return m.isPrivate, m.err
}

type mockBus struct {
	eventType  string
	payload    any
	publishErr error
}

func (m *mockBus) Publish(_ context.Context, eventType string, payload any) error {
	m.eventType = eventType
	m.payload = payload
	return m.publishErr
}

func TestFollowUserHandler_PublicUser(t *testing.T) {
	repo := &mockRepo{}
	privacy := &mockPrivacy{isPrivate: false}
	bus := &mockBus{}
	h := NewFollowUserHandler(repo, privacy, bus)

	err := h.Execute(context.Background(), FollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-2",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if bus.eventType != "follow.accepted" {
		t.Errorf("eventType = %q, want %q", bus.eventType, "follow.accepted")
	}
	f, ok := bus.payload.(*follow.Follow)
	if !ok {
		t.Fatalf("payload type = %T, want *follow.Follow", bus.payload)
	}
	if f.FollowerID != "user-1" || f.FolloweeID != "user-2" {
		t.Errorf("payload = %+v, want FollowerID=user-1 FolloweeID=user-2", f)
	}
}

func TestFollowUserHandler_PrivateUser(t *testing.T) {
	repo := &mockRepo{}
	privacy := &mockPrivacy{isPrivate: true}
	bus := &mockBus{}
	h := NewFollowUserHandler(repo, privacy, bus)

	err := h.Execute(context.Background(), FollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-2",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if bus.eventType != "follow.requested" {
		t.Errorf("eventType = %q, want %q", bus.eventType, "follow.requested")
	}
	r, ok := bus.payload.(*follow.Request)
	if !ok {
		t.Fatalf("payload type = %T, want *follow.Request", bus.payload)
	}
	if r.FollowerID != "user-1" || r.FolloweeID != "user-2" {
		t.Errorf("payload = %+v, want FollowerID=user-1 FolloweeID=user-2", r)
	}
}

func TestFollowUserHandler_SelfFollow(t *testing.T) {
	h := NewFollowUserHandler(&mockRepo{}, &mockPrivacy{}, &mockBus{})

	err := h.Execute(context.Background(), FollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-1",
	})
	if !errors.Is(err, ErrSelfFollow) {
		t.Errorf("Execute() error = %v, want %v", err, ErrSelfFollow)
	}
}

func TestFollowUserHandler_PrivacyCheckError(t *testing.T) {
	privacy := &mockPrivacy{err: errors.New("db down")}
	h := NewFollowUserHandler(&mockRepo{}, privacy, &mockBus{})

	err := h.Execute(context.Background(), FollowUserCommand{
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
	h := NewFollowUserHandler(repo, privacy, bus)

	err := h.Execute(context.Background(), FollowUserCommand{
		FollowerID: "user-1",
		TargetID:   "user-2",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if bus.eventType != "" {
		t.Errorf("event published after repo error: %q", bus.eventType)
	}
}
