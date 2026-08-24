package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type inviteFollowStub struct {
	followerID string
	followeeID string
	connected  bool
}

func (s *inviteFollowStub) AreConnected(_ context.Context, a, b string) (bool, error) {
	return s.connected && s.followerID == a && s.followeeID == b, nil
}

type inviteRepoStub struct {
	group.Repository

	inviterRole     group.Role
	inviteeIsMember bool
	inviteeInvited  bool
}

type inviteUsersStub struct{}

type inviteBusStub struct{}

func (*inviteUsersStub) GetByID(_ context.Context, _ string) (*user.User, error) {
	return &user.User{ID: "u1", Nickname: "u1-name"}, nil
}

func (s *inviteUsersStub) Create(_ context.Context, _ *user.User) error { return nil }
func (s *inviteUsersStub) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}

func (s *inviteUsersStub) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}
func (s *inviteUsersStub) Update(_ context.Context, _ *user.User) error            { return nil }
func (s *inviteUsersStub) TogglePrivacy(_ context.Context, _ string, _ bool) error { return nil }
func (s *inviteUsersStub) ListAll(_ context.Context) ([]user.User, error)          { return nil, nil }

func (*inviteBusStub) Publish(_ string, _ string, _ []byte) error { return nil }

func (*inviteBusStub) InitTopology(_ context.Context) error { return nil }

func (s *inviteBusStub) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}

func (s *inviteRepoStub) GetGroupByID(_ context.Context, _ string) (*group.Group, error) {
	return &group.Group{ID: "g1", Title: "G"}, nil
}

func (s *inviteRepoStub) GetMemberRole(_ context.Context, _, _ string) (group.Role, error) {
	return s.inviterRole, nil
}

func (s *inviteRepoStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return s.inviteeIsMember, nil
}

func (s *inviteRepoStub) IsInvited(_ context.Context, _, _ string) (bool, error) {
	return s.inviteeInvited, nil
}

func (s *inviteRepoStub) CreateInvitation(_ context.Context, _ *group.Invitation) error { return nil }

func TestInviteMember_InviteeMustFollowInviter(t *testing.T) {
	repo := &inviteRepoStub{inviterRole: group.RoleCreator}
	users := &inviteUsersStub{}
	bus := &inviteBusStub{}

	// follows table: invitee u2 follows inviter u1.
	follows := &inviteFollowStub{followerID: "u2", followeeID: "u1", connected: true}
	h := NewInviteMemberHandler(repo, follows, bus, users)

	if _, err := h.Execute(context.Background(), InviteMemberCommand{GroupID: "g1", InviterID: "u1", InviteeID: "u2"}); err != nil {
		t.Fatalf("Execute() error = %v, want success when invitee follows inviter", err)
	}

	// Inverted edge (u1 follows u2) must NOT satisfy the gate.
	inverted := &inviteFollowStub{followerID: "u1", followeeID: "u2", connected: true}
	h = NewInviteMemberHandler(repo, inverted, bus, users)

	_, err := h.Execute(context.Background(), InviteMemberCommand{GroupID: "g1", InviterID: "u1", InviteeID: "u2"})
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Execute() error = %v, want ErrNotConnected when only inviter follows invitee", err)
	}
}
