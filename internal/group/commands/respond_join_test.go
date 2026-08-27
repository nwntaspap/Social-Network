package commands

import (
	"context"
	"encoding/json"
	"testing"

	"social-network/internal/group"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type respondJoinRepoStub struct {
	*deleteGroupStub
}

func (s *respondJoinRepoStub) GetJoinRequestByID(_ context.Context, id string) (*group.JoinRequest, error) {
	return &group.JoinRequest{ID: id, GroupID: "group-1", RequesterID: "requester-1"}, nil
}

func (s *respondJoinRepoStub) GetJoinRequest(_ context.Context, groupID, requesterID string) (*group.JoinRequest, error) {
	return &group.JoinRequest{ID: "jr-1", GroupID: groupID, RequesterID: requesterID}, nil
}

func (s *respondJoinRepoStub) DeleteJoinRequest(_ context.Context, _, _ string) error { return nil }

func (s *respondJoinRepoStub) AddMember(_ context.Context, _, _ string, _ group.Role) error {
	return nil
}

type respondJoinUserStub struct{}

func (s *respondJoinUserStub) Create(_ context.Context, _ *user.User) error { return nil }

func (s *respondJoinUserStub) GetByID(_ context.Context, id string) (*user.User, error) {
	return &user.User{ID: id, Nickname: "AdminNick", AvatarPath: "/avatars/admin.png"}, nil
}

func (s *respondJoinUserStub) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}

func (s *respondJoinUserStub) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}

func (s *respondJoinUserStub) Update(_ context.Context, _ *user.User) error { return nil }

func (s *respondJoinUserStub) TogglePrivacy(_ context.Context, _ string, _ bool) error { return nil }

func (s *respondJoinUserStub) ListAll(_ context.Context) ([]user.User, error) { return nil, nil }

type respondJoinBusStub struct {
	published []eventbus.Notification
}

func (b *respondJoinBusStub) Publish(_ string, _ string, body []byte) error {
	var n eventbus.Notification
	if err := json.Unmarshal(body, &n); err != nil {
		return err
	}
	b.published = append(b.published, n)
	return nil
}

func (b *respondJoinBusStub) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}

func (b *respondJoinBusStub) InitTopology(_ context.Context) error { return nil }

// Mirrors the HTTP path: only RequestID, AdminID and Accept are set —
// GroupID/RequesterID must be resolved from the join request.
func TestRespondJoinHandler_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("accepts via request ID and notifies requester without panic", func(t *testing.T) {
		bus := &respondJoinBusStub{}
		handler := NewRespondJoinHandler(&respondJoinRepoStub{&deleteGroupStub{}}, bus, &respondJoinUserStub{})

		err := handler.Execute(ctx, RespondJoinCommand{
			RequestID: "jr-1",
			AdminID:   "creator-1",
			Accept:    true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(bus.published) != 1 {
			t.Fatalf("expected 1 published notification, got %d", len(bus.published))
		}
		n := bus.published[0]
		if n.Type != eventbus.EventGroupJoinAccepted {
			t.Errorf("Type = %q, want %q", n.Type, eventbus.EventGroupJoinAccepted)
		}
		if n.RecipientID != "requester-1" {
			t.Errorf("RecipientID = %q, want requester-1", n.RecipientID)
		}
		if n.ResourceID != "group-1" {
			t.Errorf("ResourceID = %q, want group-1", n.ResourceID)
		}
		if n.JoinRequestID != "jr-1" {
			t.Errorf("JoinRequestID = %q, want jr-1", n.JoinRequestID)
		}
		if n.ActorName != "AdminNick" {
			t.Errorf("ActorName = %q, want AdminNick", n.ActorName)
		}
	})

	t.Run("declines via request ID without adding a member", func(t *testing.T) {
		bus := &respondJoinBusStub{}
		handler := NewRespondJoinHandler(&respondJoinRepoStub{&deleteGroupStub{}}, bus, &respondJoinUserStub{})

		err := handler.Execute(ctx, RespondJoinCommand{
			RequestID: "jr-1",
			AdminID:   "creator-1",
			Accept:    false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(bus.published) != 1 {
			t.Fatalf("expected 1 published notification, got %d", len(bus.published))
		}
		if bus.published[0].Type != eventbus.EventGroupJoinDeclined {
			t.Errorf("Type = %q, want %q", bus.published[0].Type, eventbus.EventGroupJoinDeclined)
		}
	})

	t.Run("returns error when admin ID is empty", func(t *testing.T) {
		handler := NewRespondJoinHandler(&respondJoinRepoStub{&deleteGroupStub{}}, &respondJoinBusStub{}, &respondJoinUserStub{})
		err := handler.Execute(ctx, RespondJoinCommand{
			RequestID: "jr-1",
			AdminID:   "",
			Accept:    true,
		})
		if err == nil {
			t.Error("expected error for empty admin ID")
		}
	})
}
