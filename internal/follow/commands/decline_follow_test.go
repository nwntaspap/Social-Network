package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/follow"
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

func TestDeclineRequestHandler_Success(t *testing.T) {
	repo := &declineMockRepo{}
	bus := &mockBus{}
	h := NewDeclineRequestHandler(repo, bus)

	err := h.Execute(context.Background(), DeclineRequestCommand{
		FollowerID: "user-1",
		FolloweeID: "user-2",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if bus.eventType != "follow.declined" {
		t.Errorf("eventType = %q, want %q", bus.eventType, "follow.declined")
	}
	r, ok := bus.payload.(*follow.Request)
	if !ok {
		t.Fatalf("payload type = %T, want *follow.Request", bus.payload)
	}
	if r.FollowerID != "user-1" || r.FolloweeID != "user-2" {
		t.Errorf("payload = %+v, want FollowerID=user-1 FolloweeID=user-2", r)
	}
}

func TestDeclineRequestHandler_DeleteRequestError(t *testing.T) {
	repo := &declineMockRepo{deleteFollowRequestErr: errors.New("delete failed")}
	bus := &mockBus{}
	h := NewDeclineRequestHandler(repo, bus)

	err := h.Execute(context.Background(), DeclineRequestCommand{
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
