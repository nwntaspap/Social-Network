package commands

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"social-network/internal/follow"
	"social-network/internal/platform/eventbus"
)

type acceptMockRepo struct {
	deleteFollowRequestErr error
	createFollowErr        error
}

func (m *acceptMockRepo) CreateFollow(_ context.Context, _ *follow.Follow) error {
	return m.createFollowErr
}
func (m *acceptMockRepo) DeleteFollow(_ context.Context, _, _ string) error { return nil }
func (m *acceptMockRepo) GetFollowers(_ context.Context, _ string) ([]follow.Follow, error) {
	return nil, nil
}

func (m *acceptMockRepo) GetFollowing(_ context.Context, _ string) ([]follow.Follow, error) {
	return nil, nil
}

func (m *acceptMockRepo) CreateFollowRequest(_ context.Context, _ *follow.Request) error {
	return nil
}

func (m *acceptMockRepo) DeleteFollowRequest(_ context.Context, _, _ string) error {
	return m.deleteFollowRequestErr
}

func (m *acceptMockRepo) GetPendingRequests(_ context.Context, _ string) ([]follow.Request, error) {
	return nil, nil
}

func (m *acceptMockRepo) AreConnected(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (m *acceptMockRepo) GetFollowerCount(_ context.Context, _ string) (int, error) { return 0, nil }

func (m *acceptMockRepo) GetFollowingCount(_ context.Context, _ string) (int, error) { return 0, nil }

func TestAcceptRequestHandler_Success(t *testing.T) {
	repo := &acceptMockRepo{}
	bus := &mockBus{}
	h := NewAcceptRequestHandler(repo, bus, &mockUserRepo{})

	err := h.Execute(context.Background(), AcceptRequestCommand{
		FollowerID: "user-1",
		FolloweeID: "user-2",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if bus.routingKey != "follow.accepted" {
		t.Errorf("routingKey = %q, want %q", bus.routingKey, "follow.accepted")
	}
	var env eventbus.Envelope
	if err := json.Unmarshal(bus.body, &env); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if env.RecipientID != "user-1" || env.ActorID != "user-2" {
		t.Errorf("payload = %+v, want RecipientID=user-1 ActorID=user-2", env)
	}
}

func TestAcceptRequestHandler_DeleteRequestError(t *testing.T) {
	repo := &acceptMockRepo{deleteFollowRequestErr: errors.New("delete failed")}
	bus := &mockBus{}
	h := NewAcceptRequestHandler(repo, bus, &mockUserRepo{})

	err := h.Execute(context.Background(), AcceptRequestCommand{
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

func TestAcceptRequestHandler_CreateFollowError(t *testing.T) {
	repo := &acceptMockRepo{createFollowErr: errors.New("insert failed")}
	bus := &mockBus{}
	h := NewAcceptRequestHandler(repo, bus, &mockUserRepo{})

	err := h.Execute(context.Background(), AcceptRequestCommand{
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
