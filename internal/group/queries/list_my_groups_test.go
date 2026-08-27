package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

type myGroupsStub struct {
	groups   []group.Group
	total    int
	members  map[string]bool
	unread   map[string]int
	unreadID string
	err      error
}

func (s *myGroupsStub) ListUserGroups(_ context.Context, _ string, _, _ int) ([]group.Group, int, error) {
	if s.err != nil {
		return nil, 0, s.err
	}
	return s.groups, s.total, nil
}

func (s *myGroupsStub) CountGroupUnread(_ context.Context, groupIDs []string, userID string) (map[string]int, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.unreadID = userID
	_ = groupIDs
	return s.unread, nil
}

func (s *myGroupsStub) IsMember(_ context.Context, groupID, _ string) (bool, error) {
	return s.members[groupID], nil
}

func (s *myGroupsStub) IsInvited(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (s *myGroupsStub) HasPendingRequest(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func TestListMyGroupsResolver_ReturnsGroupsWithMembershipStatus(t *testing.T) {
	groups := []group.Group{{ID: "g1", Title: "Go"}, {ID: "g2", Title: "Rust"}}
	stub := &myGroupsStub{
		groups:  groups,
		total:   2,
		members: map[string]bool{"g1": true, "g2": true},
		unread:  map[string]int{"g1": 3, "g2": 0},
	}
	r := NewListMyGroupsResolver(stub)

	res, err := r.Resolve(context.Background(), ListMyGroupsQuery{UserID: "u1", Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Total != 2 || len(res.Groups) != 2 {
		t.Fatalf("Resolve() = %d groups (total %d), want 2 (2)", len(res.Groups), res.Total)
	}
	if stub.unreadID != "u1" {
		t.Errorf("CountGroupUnread called with userID %q, want %q", stub.unreadID, "u1")
	}
	for _, g := range res.Groups {
		if g.MembershipStatus != "member" {
			t.Errorf("group %s membershipStatus = %q, want %q", g.ID, g.MembershipStatus, "member")
		}
	}
	if got := res.Groups[0].UnreadCount; got != 3 {
		t.Errorf("group g1 unreadCount = %d, want 3", got)
	}
	if got := res.Groups[1].UnreadCount; got != 0 {
		t.Errorf("group g2 unreadCount = %d, want 0", got)
	}
}

func TestListMyGroupsResolver_PropagatesError(t *testing.T) {
	r := NewListMyGroupsResolver(&myGroupsStub{err: errors.New("db down")})
	if _, err := r.Resolve(context.Background(), ListMyGroupsQuery{UserID: "u1"}); err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
