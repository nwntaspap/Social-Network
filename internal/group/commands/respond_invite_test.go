package commands

import (
	"context"
	"errors"
	"testing"

	"social-network/internal/group"
)

type respondInviteStub struct {
	invitation     *group.Invitation
	getInvErr      error
	inviterRole    group.Role
	memberRole     group.Role
	deleted        bool
	addMember      bool
	createdRequest bool
	isMemberResult bool
	hasPending     bool
}

func (s *respondInviteStub) CreateGroup(_ context.Context, _ *group.Group) error { return nil }
func (s *respondInviteStub) GetGroupByID(_ context.Context, _ string) (*group.Group, error) {
	return &group.Group{ID: "group-1", CreatorID: "creator-1"}, nil
}

func (s *respondInviteStub) ListGroups(_ context.Context, _, _ int) ([]group.Group, int, error) {
	return nil, 0, nil
}
func (s *respondInviteStub) UpdateGroup(_ context.Context, _, _, _ string) error { return nil }
func (s *respondInviteStub) DeleteGroup(_ context.Context, _ string) error       { return nil }

func (s *respondInviteStub) AddMember(_ context.Context, _, _ string, role group.Role) error {
	s.addMember = true
	s.memberRole = role
	return nil
}
func (s *respondInviteStub) RemoveMember(_ context.Context, _, _ string) error { return nil }
func (s *respondInviteStub) IsMember(_ context.Context, _, _ string) (bool, error) {
	return s.isMemberResult, nil
}

func (s *respondInviteStub) GetMemberRole(_ context.Context, _, _ string) (group.Role, error) {
	if s.inviterRole != "" {
		return s.inviterRole, nil
	}
	return group.RoleMember, nil
}

func (s *respondInviteStub) GetGroupAdmins(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (s *respondInviteStub) CountMembers(_ context.Context, _ string) (int, error) { return 0, nil }
func (s *respondInviteStub) GetGroupMembers(_ context.Context, _ string, _, _ int) ([]group.Member, int, error) {
	return nil, 0, nil
}

func (s *respondInviteStub) CreateInvitation(_ context.Context, _ *group.Invitation) error {
	return nil
}

func (s *respondInviteStub) DeleteInvitation(_ context.Context, _, _ string) error {
	s.deleted = true
	return nil
}

func (s *respondInviteStub) GetInvitation(_ context.Context, _, _ string) (*group.Invitation, error) {
	if s.getInvErr != nil {
		return nil, s.getInvErr
	}
	if s.invitation == nil {
		return nil, group.ErrInvitationNotFound
	}
	return s.invitation, nil
}

func (s *respondInviteStub) IsInvited(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (s *respondInviteStub) GetPendingInvitations(_ context.Context, _ string) ([]group.Invitation, error) {
	return nil, nil
}

func (s *respondInviteStub) CreateJoinRequest(_ context.Context, _ *group.JoinRequest) error {
	s.createdRequest = true
	return nil
}
func (s *respondInviteStub) DeleteJoinRequest(_ context.Context, _, _ string) error { return nil }
func (s *respondInviteStub) GetJoinRequest(_ context.Context, _, _ string) (*group.JoinRequest, error) {
	return nil, group.ErrJoinRequestNotFound
}

func (s *respondInviteStub) GetJoinRequestByID(_ context.Context, _ string) (*group.JoinRequest, error) {
	return nil, group.ErrJoinRequestNotFound
}

func (s *respondInviteStub) HasPendingRequest(_ context.Context, _, _ string) (bool, error) {
	return s.hasPending, nil
}

func (s *respondInviteStub) GetPendingJoinRequests(_ context.Context, _ string) ([]group.JoinRequest, error) {
	return nil, nil
}

func (s *respondInviteStub) CreatePost(_ context.Context, _ *group.Post) error { return nil }
func (s *respondInviteStub) GetPostsByGroupID(_ context.Context, _ string, _ string, _, _ int) ([]group.Post, int, error) {
	return nil, 0, nil
}
func (s *respondInviteStub) CastPostVote(_ context.Context, _, _ string, _ int) error { return nil }
func (s *respondInviteStub) GetPostVoteCounts(_ context.Context, _ string) (*group.VoteCounts, error) {
	return &group.VoteCounts{}, nil
}

func (s *respondInviteStub) CreatePostComment(_ context.Context, _ *group.PostComment) error {
	return nil
}

func (s *respondInviteStub) GetPostComments(_ context.Context, _ string, _, _ int) ([]group.PostComment, int, error) {
	return nil, 0, nil
}

func (s *respondInviteStub) CountPostComments(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (s *respondInviteStub) SendGroupChatMessage(_ context.Context, _ *group.ChatMessage) error {
	return nil
}

func (s *respondInviteStub) GetGroupChatMessages(_ context.Context, _ string, _ int) ([]group.ChatMessage, error) {
	return nil, nil
}

func (s *respondInviteStub) ListGroupMemberIDs(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func TestRespondInviteHandler_Validation(t *testing.T) {
	ctx := context.Background()
	handler := NewRespondInviteHandler(&respondInviteStub{}, &mockBus{})

	t.Run("returns error when group ID is empty", func(t *testing.T) {
		_, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "", InviteeID: "user-1"})
		if err == nil {
			t.Error("expected error for empty group ID")
		}
	})

	t.Run("returns error when invitee ID is empty", func(t *testing.T) {
		_, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "group-1", InviteeID: ""})
		if err == nil {
			t.Error("expected error for empty invitee ID")
		}
	})
}

func TestRespondInviteHandler_InvitationNotFound(t *testing.T) {
	ctx := context.Background()
	handler := NewRespondInviteHandler(&respondInviteStub{getInvErr: group.ErrInvitationNotFound}, &mockBus{})
	_, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "group-1", InviteeID: "user-1", Accept: true})
	if !errors.Is(err, group.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestRespondInviteHandler_CreatorAcceptAddsMember(t *testing.T) {
	ctx := context.Background()
	existing := &group.Invitation{ID: "inv-1", GroupID: "group-1", InviterID: "creator-1", InviteeID: "user-1"}
	stub := &respondInviteStub{invitation: existing, inviterRole: group.RoleCreator}
	handler := NewRespondInviteHandler(stub, &mockBus{})

	result, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "group-1", InviteeID: "user-1", Accept: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != RespondInviteMember {
		t.Errorf("result = %q, want %q", result, RespondInviteMember)
	}
	if !stub.deleted {
		t.Error("expected invitation to be deleted")
	}
	if !stub.addMember {
		t.Error("expected AddMember to be called")
	}
	if stub.memberRole != group.RoleMember {
		t.Errorf("role = %q, want %q", stub.memberRole, group.RoleMember)
	}
	if stub.createdRequest {
		t.Error("expected no join request for creator invite")
	}
}

func TestRespondInviteHandler_MemberAcceptCreatesPendingRequest(t *testing.T) {
	ctx := context.Background()
	existing := &group.Invitation{ID: "inv-1", GroupID: "group-1", InviterID: "creator-1", InviteeID: "user-1"}
	stub := &respondInviteStub{invitation: existing}
	handler := NewRespondInviteHandler(stub, &mockBus{})

	result, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "group-1", InviteeID: "user-1", Accept: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != RespondInvitePending {
		t.Errorf("result = %q, want %q", result, RespondInvitePending)
	}
	if !stub.deleted {
		t.Error("expected invitation to be deleted")
	}
	if stub.addMember {
		t.Error("expected AddMember NOT to be called for member invite")
	}
	if !stub.createdRequest {
		t.Error("expected CreateJoinRequest to be called")
	}
}

func TestRespondInviteHandler_MemberAcceptSkipsDuplicateRequest(t *testing.T) {
	ctx := context.Background()
	existing := &group.Invitation{ID: "inv-1", GroupID: "group-1", InviterID: "creator-1", InviteeID: "user-1"}
	stub := &respondInviteStub{invitation: existing, hasPending: true}
	handler := NewRespondInviteHandler(stub, &mockBus{})

	result, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "group-1", InviteeID: "user-1", Accept: true})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != RespondInvitePending {
		t.Errorf("result = %q, want %q", result, RespondInvitePending)
	}
	if stub.createdRequest {
		t.Error("expected no duplicate join request")
	}
}

func TestRespondInviteHandler_DeclineDeletesInvitation(t *testing.T) {
	ctx := context.Background()
	existing := &group.Invitation{ID: "inv-1", GroupID: "group-1", InviterID: "creator-1", InviteeID: "user-1"}
	stub := &respondInviteStub{invitation: existing}
	handler := NewRespondInviteHandler(stub, &mockBus{})

	_, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "group-1", InviteeID: "user-1", Accept: false})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !stub.deleted {
		t.Error("expected invitation to be deleted")
	}
	if stub.addMember {
		t.Error("expected AddMember NOT to be called on decline")
	}
	if stub.createdRequest {
		t.Error("expected no join request on decline")
	}
}
