package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-network/internal/group"
)

type pendingJoinRequestsStub struct {
	role        group.Role
	memberErr   error
	requests    []group.JoinRequest
	groupByID   *group.Group
	getGroupErr error
}

func (s *pendingJoinRequestsStub) GetMemberRole(_ context.Context, _, _ string) (group.Role, error) {
	if s.memberErr != nil {
		return "", s.memberErr
	}
	return s.role, nil
}

func (s *pendingJoinRequestsStub) GetPendingJoinRequests(_ context.Context, _ string) ([]group.JoinRequest, error) {
	return s.requests, nil
}

func (s *pendingJoinRequestsStub) GetGroupByID(_ context.Context, _ string) (*group.Group, error) {
	if s.getGroupErr != nil {
		return nil, s.getGroupErr
	}
	return s.groupByID, nil
}

func TestGetPendingJoinRequestsResolver_EnrichesWithGroup(t *testing.T) {
	ctx := context.Background()
	jr := group.JoinRequest{ID: "jr1", GroupID: "g1", RequesterID: "u3", CreatedAt: time.Now()}
	resolver := NewGetPendingJoinRequestsResolver(&pendingJoinRequestsStub{
		role:      group.RoleCreator,
		requests:  []group.JoinRequest{jr},
		groupByID: &group.Group{ID: "g1", Title: "Go Meetup", CreatorID: "u1"},
	})

	res, err := resolver.Resolve(ctx, GetPendingJoinRequestsQuery{GroupID: "g1", UserID: "u1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(res.Requests) != 1 {
		t.Fatalf("len(requests) = %d, want 1", len(res.Requests))
	}
	r := res.Requests[0]
	if r.Request.ID != "jr1" || r.Request.RequesterID != "u3" {
		t.Errorf("request = %+v, want jr1/u3", r.Request)
	}
	if r.Group.ID != "g1" || r.Group.Title != "Go Meetup" {
		t.Errorf("group = %+v, want g1/Go Meetup", r.Group)
	}
}

func TestGetPendingJoinRequestsResolver_AdminAllowed(t *testing.T) {
	ctx := context.Background()
	resolver := NewGetPendingJoinRequestsResolver(&pendingJoinRequestsStub{
		role:      group.RoleAdmin,
		groupByID: &group.Group{ID: "g1"},
	})

	res, err := resolver.Resolve(ctx, GetPendingJoinRequestsQuery{GroupID: "g1", UserID: "u2"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result for admin")
	}
}

func TestGetPendingJoinRequestsResolver_RejectsNonAdmin(t *testing.T) {
	ctx := context.Background()
	resolver := NewGetPendingJoinRequestsResolver(&pendingJoinRequestsStub{role: group.RoleMember})

	_, err := resolver.Resolve(ctx, GetPendingJoinRequestsQuery{GroupID: "g1", UserID: "u3"})
	if !errors.Is(err, group.ErrNotAdmin) {
		t.Fatalf("Resolve() error = %v, want ErrNotAdmin", err)
	}
}

func TestGetPendingJoinRequestsResolver_RejectsNonMember(t *testing.T) {
	ctx := context.Background()
	resolver := NewGetPendingJoinRequestsResolver(&pendingJoinRequestsStub{
		memberErr: group.ErrNotMember,
	})

	_, err := resolver.Resolve(ctx, GetPendingJoinRequestsQuery{GroupID: "g1", UserID: "u9"})
	if !errors.Is(err, group.ErrNotAdmin) {
		t.Fatalf("Resolve() error = %v, want ErrNotAdmin", err)
	}
}

func TestGetPendingJoinRequestsResolver_Empty(t *testing.T) {
	ctx := context.Background()
	resolver := NewGetPendingJoinRequestsResolver(&pendingJoinRequestsStub{
		role:      group.RoleCreator,
		groupByID: &group.Group{ID: "g1"},
	})

	res, err := resolver.Resolve(ctx, GetPendingJoinRequestsQuery{GroupID: "g1", UserID: "u1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(res.Requests) != 0 {
		t.Errorf("len(requests) = %d, want 0", len(res.Requests))
	}
}

func TestGetPendingJoinRequestsResolver_MissingFields(t *testing.T) {
	ctx := context.Background()
	resolver := NewGetPendingJoinRequestsResolver(&pendingJoinRequestsStub{})

	if _, err := resolver.Resolve(ctx, GetPendingJoinRequestsQuery{}); err == nil {
		t.Error("expected error when group_id and user_id are missing")
	}
}
