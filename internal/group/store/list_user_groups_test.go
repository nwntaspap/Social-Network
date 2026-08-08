package store

import (
	"context"
	"testing"

	"social-network/internal/group"
)

func TestListUserGroups_ReturnsOnlyMemberGroupsNewestFirst(t *testing.T) {
	s := setupGroupChatStore(t)
	ctx := context.Background()

	seedGroupAt(t, s, &group.Group{ID: "g1", Title: "Go", CreatorID: "u1"}, "2025-01-01 10:00:00")
	seedGroupAt(t, s, &group.Group{ID: "g2", Title: "Rust", CreatorID: "u1"}, "2025-01-02 10:00:00")
	seedGroupAt(t, s, &group.Group{ID: "g3", Title: "Zig", CreatorID: "u1"}, "2025-01-03 10:00:00")

	for _, m := range [][2]string{{"g1", "u1"}, {"g2", "u1"}, {"g2", "u2"}} {
		if err := s.AddMember(ctx, m[0], m[1], group.RoleMember); err != nil {
			t.Fatalf("add member %v: %v", m, err)
		}
	}

	groups, total, err := s.ListUserGroups(ctx, "u1", 1, 10)
	if err != nil {
		t.Fatalf("ListUserGroups(u1) error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(groups) != 2 {
		t.Fatalf("ListUserGroups(u1) returned %d groups, want 2", len(groups))
	}

	if groups[0].ID != "g2" || groups[1].ID != "g1" {
		t.Errorf("ListUserGroups(u1) order = [%s, %s], want [g2, g1] (newest first)", groups[0].ID, groups[1].ID)
	}

	only, total, err := s.ListUserGroups(ctx, "u2", 1, 10)
	if err != nil {
		t.Fatalf("ListUserGroups(u2) error = %v", err)
	}
	if total != 1 || len(only) != 1 || only[0].ID != "g2" {
		t.Errorf("ListUserGroups(u2) = %+v (total %d), want [g2] (1)", only, total)
	}

	none, total, err := s.ListUserGroups(ctx, "nobody", 1, 10)
	if err != nil {
		t.Fatalf("ListUserGroups(nobody) error = %v", err)
	}
	if total != 0 || len(none) != 0 {
		t.Errorf("ListUserGroups(nobody) = %+v (total %d), want empty (0)", none, total)
	}
}

func TestListUserGroups_RespectsPagination(t *testing.T) {
	s := setupGroupChatStore(t)
	ctx := context.Background()

	for i := range 3 {
		id := "g" + string(rune('1'+i))
		seedGroupAt(t, s, &group.Group{ID: id, Title: "Group " + id, CreatorID: "u1"}, "2025-01-0"+string(rune('1'+i))+" 10:00:00")
		if err := s.AddMember(ctx, id, "u1", group.RoleMember); err != nil {
			t.Fatalf("add member %s: %v", id, err)
		}
	}

	groups, total, err := s.ListUserGroups(ctx, "u1", 2, 2)
	if err != nil {
		t.Fatalf("ListUserGroups page 2 error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(groups) != 1 {
		t.Errorf("page 2 returned %d groups, want 1", len(groups))
	}
}
