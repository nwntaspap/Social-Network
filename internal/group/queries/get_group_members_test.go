package queries

import (
	"context"
	"testing"
	"time"

	"social-network/internal/group"
)

type getGroupMembersStub struct {
	getGroupMembersFn func(ctx context.Context, groupID string, page, size int) ([]group.Member, int, error)
}

func (s *getGroupMembersStub) AddMember(_ context.Context, _, _ string, _ group.Role) error {
	return nil
}
func (s *getGroupMembersStub) RemoveMember(_ context.Context, _, _ string) error { return nil }
func (s *getGroupMembersStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

func (s *getGroupMembersStub) GetMemberRole(_ context.Context, _, _ string) (group.Role, error) {
	return group.RoleMember, nil
}

func (s *getGroupMembersStub) CountMembers(_ context.Context, _ string) (int, error) { return 0, nil }

func (s *getGroupMembersStub) GetGroupMembers(ctx context.Context, groupID string, page, size int) ([]group.Member, int, error) {
	if s.getGroupMembersFn != nil {
		return s.getGroupMembersFn(ctx, groupID, page, size)
	}
	return nil, 0, nil
}

func TestGetGroupMembersResolver_Resolve(t *testing.T) {
	ctx := context.Background()

	t.Run("returns members for group", func(t *testing.T) {
		expected := []group.Member{
			{GroupID: "g1", UserID: "u1", Role: group.RoleCreator, JoinedAt: time.Now()},
			{GroupID: "g1", UserID: "u2", Role: group.RoleMember, JoinedAt: time.Now()},
		}
		stub := &getGroupMembersStub{
			getGroupMembersFn: func(_ context.Context, _ string, _, _ int) ([]group.Member, int, error) {
				return expected, 2, nil
			},
		}
		resolver := NewGetGroupMembersResolver(stub)
		result, err := resolver.Resolve(ctx, GetGroupMembersQuery{
			GroupID: "g1",
			Page:    1,
			Size:    10,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Members) != 2 {
			t.Errorf("expected 2 members, got %d", len(result.Members))
		}
		if result.Total != 2 {
			t.Errorf("expected total 2, got %d", result.Total)
		}
	})

	t.Run("returns empty list for group with no members", func(t *testing.T) {
		stub := &getGroupMembersStub{
			getGroupMembersFn: func(_ context.Context, _ string, _, _ int) ([]group.Member, int, error) {
				return []group.Member{}, 0, nil
			},
		}
		resolver := NewGetGroupMembersResolver(stub)
		result, err := resolver.Resolve(ctx, GetGroupMembersQuery{
			GroupID: "g1",
			Page:    1,
			Size:    10,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Members) != 0 {
			t.Errorf("expected 0 members, got %d", len(result.Members))
		}
	})
}
