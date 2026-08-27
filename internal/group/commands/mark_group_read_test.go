package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

type markGroupReadRepoStub struct {
	group.Repository

	isMember  bool
	memberErr error
	markErr   error
	gotGroup  string
	gotUser   string
}

func (s *markGroupReadRepoStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return s.isMember, s.memberErr
}

func (s *markGroupReadRepoStub) MarkGroupRead(_ context.Context, groupID, userID string) error {
	if s.markErr != nil {
		return s.markErr
	}
	s.gotGroup, s.gotUser = groupID, userID
	return nil
}

func TestMarkGroupReadHandler_MarksReadForMember(t *testing.T) {
	ctx := context.Background()
	repo := &markGroupReadRepoStub{isMember: true}
	h := NewMarkGroupReadHandler(repo)

	if err := h.Execute(ctx, MarkGroupReadCommand{GroupID: "g1", UserID: "u1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.gotGroup != "g1" || repo.gotUser != "u1" {
		t.Fatalf("MarkGroupRead(%s, %s), want (g1, u1)", repo.gotGroup, repo.gotUser)
	}
}

func TestMarkGroupReadHandler_RejectsNonMember(t *testing.T) {
	ctx := context.Background()
	repo := &markGroupReadRepoStub{isMember: false}
	h := NewMarkGroupReadHandler(repo)

	if err := h.Execute(ctx, MarkGroupReadCommand{GroupID: "g1", UserID: "u1"}); !errors.Is(err, group.ErrNotMember) {
		t.Fatalf("error = %v, want group.ErrNotMember", err)
	}
}

func TestMarkGroupReadHandler_RequiresFields(t *testing.T) {
	ctx := context.Background()
	repo := &markGroupReadRepoStub{isMember: true}
	h := NewMarkGroupReadHandler(repo)

	for _, cmd := range []MarkGroupReadCommand{
		{GroupID: "", UserID: "u1"},
		{GroupID: "g1", UserID: ""},
	} {
		if err := h.Execute(ctx, cmd); err == nil {
			t.Fatalf("Execute(%+v) expected error", cmd)
		}
	}
}
