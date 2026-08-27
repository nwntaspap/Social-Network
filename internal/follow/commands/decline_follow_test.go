package commands

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"social-network/internal/follow"
	"social-network/internal/platform/eventbus"
)

type declineMockRepo struct {
	deleteFollowRequestErr error
}

func (m *declineMockRepo) CreateFollow(_ context.Context, _ *follow.Follow) error {
	return nil
}
func (m *declineMockRepo) DeleteFollow(_ context.Context, _, _ string) error { return nil }
func (m *declineMockRepo) GetFollowers(_ context.Context, _ string) ([]follow.Follow, error) {
	return nil, nil
}

func (m *declineMockRepo) GetFollowing(_ context.Context, _ string) ([]follow.Follow, error) {
	return nil, nil
}

func (m *declineMockRepo) CreateFollowRequest(_ context.Context, _ *follow.Request) error {
	return nil
}

func (m *declineMockRepo) DeleteFollowRequest(_ context.Context, _, _ string) error {
	return m.deleteFollowRequestErr
}

func (m *declineMockRepo) GetPendingRequests(_ context.Context, _ string) ([]follow.Request, error) {
	return nil, nil
}

func (m *declineMockRepo) AreConnected(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (m *declineMockRepo) GetFollowerCount(_ context.Context, _ string) (int, error) { return 0, nil }

func (m *declineMockRepo) GetFollowingCount(_ context.Context, _ string) (int, error) { return 0, nil }

func TestDeclineRequestHandler_Success(t *testing.T) {
	repo := &declineMockRepo{}
	bus := &mockBus{}
	h := NewDeclineRequestHandler(repo, bus, &mockUserRepo{})

	err := h.Execute(context.Background(), DeclineRequestCommand{
		FollowerID: "user-1",
		FolloweeID: "user-2",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if bus.routingKey != "deleted" {
		t.Errorf("routingKey = %q, want %q", bus.routingKey, "deleted")
	}
	var env eventbus.Notification
	if err := json.Unmarshal(bus.body, &env); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if env.RecipientID != "user-1" || env.ActorID != "user-2" {
		t.Errorf("payload = %+v, want RecipientID=user-1 ActorID=user-2", env)
	}
}

func TestDeclineRequestHandler_DeleteRequestError(t *testing.T) {
	repo := &declineMockRepo{deleteFollowRequestErr: errors.New("delete failed")}
	bus := &mockBus{}
	h := NewDeclineRequestHandler(repo, bus, &mockUserRepo{})

	err := h.Execute(context.Background(), DeclineRequestCommand{
		FollowerID: "user-1",
		FolloweeID: "user-2",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if bus.routingKey != "" {
		t.Errorf("event published after repo error: %q", bus.routingKey)
	}
}
