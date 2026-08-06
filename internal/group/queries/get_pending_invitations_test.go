package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/group"
)

type pendingInvitationsStub struct {
	invitations []group.Invitation
	getGroupErr error
}

func (s *pendingInvitationsStub) GetPendingInvitations(_ context.Context, _ string) ([]group.Invitation, error) {
	return s.invitations, nil
}

func (s *pendingInvitationsStub) GetGroupByID(_ context.Context, _ string) (*group.Group, error) {
	if s.getGroupErr != nil {
		return nil, s.getGroupErr
	}
	return &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u2"}, nil
}

func TestGetPendingInvitationsResolver_EnrichesWithGroup(t *testing.T) {
	ctx := context.Background()
	inv := group.Invitation{ID: "i1", GroupID: "g1", InviterID: "u2", InviteeID: "u1", CreatedAt: time.Now()}
	resolver := NewGetPendingInvitationsResolver(&pendingInvitationsStub{invitations: []group.Invitation{inv}})

	res, err := resolver.Resolve(ctx, GetPendingInvitationsQuery{UserID: "u1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(res.Invitations) != 1 {
		t.Fatalf("len(invitations) = %d, want 1", len(res.Invitations))
	}
	it := res.Invitations[0]
	if it.Invitation.ID != "i1" {
		t.Errorf("invitation id = %q, want %q", it.Invitation.ID, "i1")
	}
	if it.Group.ID != "g1" || it.Group.Title != "Go Meetup" {
		t.Errorf("group = %+v, want g1/Go Meetup", it.Group)
	}
}

func TestGetPendingInvitationsResolver_Empty(t *testing.T) {
	ctx := context.Background()
	resolver := NewGetPendingInvitationsResolver(&pendingInvitationsStub{})

	res, err := resolver.Resolve(ctx, GetPendingInvitationsQuery{UserID: "u1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(res.Invitations) != 0 {
		t.Errorf("len(invitations) = %d, want 0", len(res.Invitations))
	}
}

func TestGetPendingInvitationsResolver_GroupLookupError(t *testing.T) {
	ctx := context.Background()
	inv := group.Invitation{ID: "i1", GroupID: "g1", InviteeID: "u1"}
	resolver := NewGetPendingInvitationsResolver(&pendingInvitationsStub{
		invitations: []group.Invitation{inv},
		getGroupErr: errors.New("boom"),
	})

	if _, err := resolver.Resolve(ctx, GetPendingInvitationsQuery{UserID: "u1"}); err == nil {
		t.Error("expected error when group lookup fails")
	}
}
