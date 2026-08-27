package queries

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

type groupPresenceStub struct {
	memberIDs []string
	isMember  bool
	err       error
}

func (s *groupPresenceStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.isMember, nil
}

func (s *groupPresenceStub) ListGroupMemberIDs(_ context.Context, _ string) ([]string, error) {
	return s.memberIDs, nil
}

func TestGetGroupPresenceResolver_CountsOnlineMembers(t *testing.T) {
	r := NewGetGroupPresenceResolver(
		&groupPresenceStub{isMember: true, memberIDs: []string{"u1", "u2", "u3"}},
		func(id string) bool { return id == "u1" || id == "u3" },
	)

	res, err := r.Resolve(context.Background(), GetGroupPresenceQuery{GroupID: "g1", UserID: "u1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Total != 3 {
		t.Errorf("Total = %d, want 3", res.Total)
	}
	if res.Online != 2 {
		t.Errorf("Online = %d, want 2", res.Online)
	}
	if len(res.Members) != 3 {
		t.Fatalf("Members = %d, want 3", len(res.Members))
	}
	byID := map[string]bool{}
	for _, m := range res.Members {
		byID[m.ID] = m.IsOnline
	}
	if !byID["u1"] || byID["u2"] || !byID["u3"] {
		t.Errorf("member online flags = %v, want u1/u3 online", byID)
	}
}

func TestGetGroupPresenceResolver_NilCheckerTreatsAllOffline(t *testing.T) {
	r := NewGetGroupPresenceResolver(
		&groupPresenceStub{isMember: true, memberIDs: []string{"u1", "u2"}},
		nil,
	)

	res, err := r.Resolve(context.Background(), GetGroupPresenceQuery{GroupID: "g1", UserID: "u1"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Online != 0 || res.Total != 2 {
		t.Errorf("Online/Total = %d/%d, want 0/2", res.Online, res.Total)
	}
}

func TestGetGroupPresenceResolver_RejectsNonMember(t *testing.T) {
	r := NewGetGroupPresenceResolver(
		&groupPresenceStub{isMember: false, memberIDs: []string{"u1"}},
		func(string) bool { return true },
	)

	if _, err := r.Resolve(context.Background(), GetGroupPresenceQuery{GroupID: "g1", UserID: "u9"}); !errors.Is(err, group.ErrNotMember) {
		t.Fatalf("Resolve() error = %v, want ErrNotMember", err)
	}
}
