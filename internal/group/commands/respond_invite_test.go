package commands

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"social-network/internal/group"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
)

type respondInviteStub struct {
	invitation       *group.Invitation
	getInvErr        error
	inviterRole      group.Role
	memberRole       group.Role
	deleted          bool
	addMember        bool
	createdRequest   bool
	createdRequestID string
	isMemberResult   bool
	hasPending       bool
	admins           []string
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
	return s.admins, nil
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

func (s *respondInviteStub) CreateJoinRequest(_ context.Context, jr *group.JoinRequest) error {
	s.createdRequest = true
	s.createdRequestID = jr.ID
	return nil
}
func (s *respondInviteStub) DeleteJoinRequest(_ context.Context, _, _ string) error { return nil }
func (s *respondInviteStub) GetJoinRequest(_ context.Context, _, _ string) (*group.JoinRequest, error) {
	if s.hasPending {
		return &group.JoinRequest{ID: "jr-existing", GroupID: "group-1", RequesterID: "user-1"}, nil
	}
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

func (s *respondInviteStub) GetSentInvitationInviteeIDs(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}

func (s *respondInviteStub) CreatePost(_ context.Context, _ *group.Post) error { return nil }
func (s *respondInviteStub) GetPostsByGroupID(_ context.Context, _ string, _ string, _, _ int) ([]group.Post, int, error) {
	return nil, 0, nil
}

func (s *respondInviteStub) CastPostVote(_ context.Context, _, _ string, _ int) (group.VoteChange, error) {
	return group.VoteChangeAdded, nil
}

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
	handler := NewRespondInviteHandler(&respondInviteStub{}, &mockBus{}, &userStub{})

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
	handler := NewRespondInviteHandler(&respondInviteStub{getInvErr: group.ErrInvitationNotFound}, &mockBus{}, &userStub{})
	_, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "group-1", InviteeID: "user-1", Accept: true})
	if !errors.Is(err, group.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestRespondInviteHandler_CreatorAcceptAddsMember(t *testing.T) {
	ctx := context.Background()
	existing := &group.Invitation{ID: "inv-1", GroupID: "group-1", InviterID: "creator-1", InviteeID: "user-1"}
	stub := &respondInviteStub{invitation: existing, inviterRole: group.RoleCreator}
	handler := NewRespondInviteHandler(stub, &mockBus{}, &userStub{})

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
	handler := NewRespondInviteHandler(stub, &mockBus{}, &userStub{})

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
	handler := NewRespondInviteHandler(stub, &mockBus{}, &userStub{})

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
	handler := NewRespondInviteHandler(stub, &mockBus{}, &userStub{})

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

func (s *respondInviteStub) GetPostByID(_ context.Context, _ string) (*group.Post, error) {
	return nil, group.ErrPostNotFound
}

type userStub struct{}

func (s *userStub) Create(_ context.Context, _ *user.User) error { return nil }
func (s *userStub) GetByID(_ context.Context, id string) (*user.User, error) {
	return &user.User{ID: id, Nickname: "User" + id, AvatarPath: "/avatars/" + id + ".png"}, nil
}

func (s *userStub) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}

func (s *userStub) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}
func (s *userStub) Update(_ context.Context, _ *user.User) error { return nil }
func (s *userStub) TogglePrivacy(_ context.Context, _ string, _ bool) error {
	return nil
}
func (s *userStub) ListAll(_ context.Context) ([]user.User, error) { return nil, nil }

type recordingBus struct {
	publishes []eventbus.Notification
}

func (b *recordingBus) Publish(_ string, _ string, body []byte) error {
	var n eventbus.Notification
	if err := json.Unmarshal(body, &n); err == nil {
		b.publishes = append(b.publishes, n)
	}
	return nil
}

func (b *recordingBus) Subscribe(_ context.Context, _ string) (<-chan eventbus.Message, error) {
	ch := make(chan eventbus.Message)
	close(ch)
	return ch, nil
}

func (b *recordingBus) InitTopology(_ context.Context) error { return nil }

func publishedByType(b *recordingBus, typ string) *eventbus.Notification {
	for i := range b.publishes {
		if b.publishes[i].Type == typ {
			n := b.publishes[i]
			return &n
		}
	}
	return nil
}

func TestRespondInviteHandler_MemberAcceptPublishesPendingAndJoinRequest(t *testing.T) {
	ctx := context.Background()
	existing := &group.Invitation{ID: "inv-1", GroupID: "group-1", InviterID: "member-1", InviteeID: "user-1"}
	stub := &respondInviteStub{
		invitation: existing,
		admins:     []string{"creator-1", "admin-1"},
	}
	bus := &recordingBus{}
	handler := NewRespondInviteHandler(stub, bus, &userStub{})

	result, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "group-1", InviteeID: "user-1", Accept: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != RespondInvitePending {
		t.Fatalf("result = %q, want %q", result, RespondInvitePending)
	}

	pending := publishedByType(bus, eventbus.EventGroupInviteAcceptedPending)
	if pending == nil {
		t.Fatal("expected group.invitation.accepted_pending notification")
	}
	if pending.RecipientID != "user-1" {
		t.Errorf("pending RecipientID = %q, want user-1", pending.RecipientID)
	}
	if pending.ActorID != "member-1" {
		t.Errorf("pending ActorID = %q, want member-1", pending.ActorID)
	}
	if pending.ResourceID != "group-1" {
		t.Errorf("pending ResourceID = %q, want group-1", pending.ResourceID)
	}
	if pending.JoinRequestID == "" {
		t.Error("expected pending notification to carry JoinRequestID")
	}

	requested := publishedByType(bus, eventbus.EventGroupJoinRequested)
	if requested == nil {
		t.Fatal("expected group.join.requested fan-out")
	}
	if len(requested.MultipleRecipients) != 2 {
		t.Errorf("requested recipients = %v, want creator-1 and admin-1", requested.MultipleRecipients)
	}
	if requested.ActorID != "user-1" {
		t.Errorf("requested ActorID = %q, want user-1", requested.ActorID)
	}
	if requested.ActorName != "Useruser-1" {
		t.Errorf("requested ActorName = %q, want Useruser-1", requested.ActorName)
	}
	if requested.JoinRequestID != pending.JoinRequestID {
		t.Errorf("requested JoinRequestID = %q, want %q", requested.JoinRequestID, pending.JoinRequestID)
	}
	if requested.JoinRequestID != stub.createdRequestID {
		t.Errorf("requested JoinRequestID = %q, want created request %q", requested.JoinRequestID, stub.createdRequestID)
	}
}

func TestRespondInviteHandler_MemberAcceptExistingRequestReusesID(t *testing.T) {
	ctx := context.Background()
	existing := &group.Invitation{ID: "inv-1", GroupID: "group-1", InviterID: "member-1", InviteeID: "user-1"}
	stub := &respondInviteStub{
		invitation: existing,
		hasPending: true,
		admins:     []string{"creator-1"},
	}
	bus := &recordingBus{}
	handler := NewRespondInviteHandler(stub, bus, &userStub{})

	result, err := handler.Execute(ctx, RespondInviteCommand{GroupID: "group-1", InviteeID: "user-1", Accept: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != RespondInvitePending {
		t.Fatalf("result = %q, want %q", result, RespondInvitePending)
	}
	if stub.createdRequest {
		t.Fatal("expected no duplicate join request")
	}

	pending := publishedByType(bus, eventbus.EventGroupInviteAcceptedPending)
	if pending == nil {
		t.Fatal("expected group.invitation.accepted_pending notification")
	}
	if pending.JoinRequestID != "jr-existing" {
		t.Errorf("pending JoinRequestID = %q, want jr-existing", pending.JoinRequestID)
	}

	if requested := publishedByType(bus, eventbus.EventGroupJoinRequested); requested != nil {
		t.Error("expected no duplicate group.join.requested notification")
	}
}
