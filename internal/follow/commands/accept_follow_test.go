package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/follow"
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

func TestAcceptRequestHandler_Success(t *testing.T) {
	repo := &acceptMockRepo{}
	bus := &mockBus{}
	h := NewAcceptRequestHandler(repo, bus)

	err := h.Execute(context.Background(), AcceptRequestCommand{
		FollowerID: "user-1",
		FolloweeID: "user-2",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if bus.eventType != "follow.accepted" {
		t.Errorf("eventType = %q, want %q", bus.eventType, "follow.accepted")
	}
	r, ok := bus.payload.(*follow.Request)
	if !ok {
		t.Fatalf("payload type = %T, want *follow.Request", bus.payload)
	}
	if r.FollowerID != "user-1" || r.FolloweeID != "user-2" {
		t.Errorf("payload = %+v, want FollowerID=user-1 FolloweeID=user-2", r)
	}
}

func TestAcceptRequestHandler_DeleteRequestError(t *testing.T) {
	repo := &acceptMockRepo{deleteFollowRequestErr: errors.New("delete failed")}
	bus := &mockBus{}
	h := NewAcceptRequestHandler(repo, bus)

	err := h.Execute(context.Background(), AcceptRequestCommand{
		FollowerID: "user-1",
		FolloweeID: "user-2",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if bus.eventType != "" {
		t.Errorf("event published after repo error: %q", bus.eventType)
	}
}

func TestAcceptRequestHandler_CreateFollowError(t *testing.T) {
	repo := &acceptMockRepo{createFollowErr: errors.New("insert failed")}
	bus := &mockBus{}
	h := NewAcceptRequestHandler(repo, bus)

	err := h.Execute(context.Background(), AcceptRequestCommand{
		FollowerID: "user-1",
		FolloweeID: "user-2",
	})
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if bus.eventType != "" {
		t.Errorf("event published after repo error: %q", bus.eventType)
	}
}
