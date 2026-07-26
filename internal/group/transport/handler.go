package transport

import (
	"context"
	"net/http"

	"social-network/internal/group"
	"social-network/internal/group/commands"
	"social-network/internal/group/queries"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type UserLookup interface {
	GetUserByID(ctx context.Context, id string) (*UserResult, error)
}

type CreateGroupExecutor interface {
	Execute(ctx context.Context, cmd commands.CreateGroupCommand) (*group.Group, error)
}

type InviteMemberExecutor interface {
	Execute(ctx context.Context, cmd commands.InviteMemberCommand) (*group.Invitation, error)
}

type RespondInviteExecutor interface {
	Execute(ctx context.Context, cmd commands.RespondInviteCommand) error
}

type RequestJoinExecutor interface {
	Execute(ctx context.Context, cmd commands.RequestJoinCommand) (*group.JoinRequest, error)
}

type RespondJoinExecutor interface {
	Execute(ctx context.Context, cmd commands.RespondJoinCommand) error
}

type CreateGroupPostExecutor interface {
	Execute(ctx context.Context, cmd commands.CreateGroupPostCommand) (*group.Post, error)
}

type SendGroupMessageExecutor interface {
	Execute(ctx context.Context, cmd commands.SendGroupMessageCommand) (commands.SendGroupMessageResult, error)
}

type CreateGroupPostCommentExecutor interface {
	Execute(ctx context.Context, cmd commands.CreateGroupPostCommentCommand) (*group.PostComment, error)
}

type LeaveGroupExecutor interface {
	Execute(ctx context.Context, cmd commands.LeaveGroupCommand) error
}

type UpdateGroupExecutor interface {
	Execute(ctx context.Context, cmd commands.UpdateGroupCommand) (*group.Group, error)
}

type DeleteGroupExecutor interface {
	Execute(ctx context.Context, cmd commands.DeleteGroupCommand) error
}

type ListGroupsResolver interface {
	Resolve(ctx context.Context, q queries.ListGroupsQuery) (*queries.ListGroupsResult, error)
}

type GetGroupResolver interface {
	Resolve(ctx context.Context, q queries.GetGroupQuery) (*queries.GetGroupResult, error)
}

type GetGroupFeedResolver interface {
	Resolve(ctx context.Context, q queries.GetGroupFeedQuery) (*queries.GetGroupFeedResult, error)
}

type GetGroupChatResolver interface {
	Resolve(ctx context.Context, q queries.GetGroupChatQuery) (*queries.GetGroupChatResult, error)
}

type GetGroupPostCommentsResolver interface {
	Resolve(ctx context.Context, q queries.GetGroupPostCommentsQuery) (*queries.GetGroupPostCommentsResult, error)
}

type GetGroupMembersResolver interface {
	Resolve(ctx context.Context, q queries.GetGroupMembersQuery) (*queries.GetGroupMembersResult, error)
}

type Handler struct {
	createGroup            CreateGroupExecutor
	inviteMember           InviteMemberExecutor
	respondInvite          RespondInviteExecutor
	requestJoin            RequestJoinExecutor
	respondJoin            RespondJoinExecutor
	createGroupPost        CreateGroupPostExecutor
	createGroupPostComment CreateGroupPostCommentExecutor
	leaveGroup             LeaveGroupExecutor
	updateGroup            UpdateGroupExecutor
	deleteGroup            DeleteGroupExecutor
	listGroups             ListGroupsResolver
	getGroup               GetGroupResolver
	getGroupFeed           GetGroupFeedResolver
	getGroupChat           GetGroupChatResolver
	getGroupPostComments   GetGroupPostCommentsResolver
	getGroupMembers        GetGroupMembersResolver
	userLookup             UserLookup
	extractUser            UserExtractor
}

func NewHandler(
	extractUser UserExtractor,
	userLookup UserLookup,
	createGroup CreateGroupExecutor,
	inviteMember InviteMemberExecutor,
	respondInvite RespondInviteExecutor,
	requestJoin RequestJoinExecutor,
	respondJoin RespondJoinExecutor,
	createGroupPost CreateGroupPostExecutor,
	createGroupPostComment CreateGroupPostCommentExecutor,
	leaveGroup LeaveGroupExecutor,
	updateGroup UpdateGroupExecutor,
	deleteGroup DeleteGroupExecutor,
	listGroups ListGroupsResolver,
	getGroup GetGroupResolver,
	getGroupFeed GetGroupFeedResolver,
	getGroupChat GetGroupChatResolver,
	getGroupPostComments GetGroupPostCommentsResolver,
	getGroupMembers GetGroupMembersResolver,
) *Handler {
	return &Handler{
		createGroup:            createGroup,
		inviteMember:           inviteMember,
		respondInvite:          respondInvite,
		requestJoin:            requestJoin,
		respondJoin:            respondJoin,
		createGroupPost:        createGroupPost,
		createGroupPostComment: createGroupPostComment,
		leaveGroup:             leaveGroup,
		updateGroup:            updateGroup,
		deleteGroup:            deleteGroup,
		listGroups:             listGroups,
		getGroup:               getGroup,
		getGroupFeed:           getGroupFeed,
		getGroupChat:           getGroupChat,
		getGroupPostComments:   getGroupPostComments,
		getGroupMembers:        getGroupMembers,
		userLookup:             userLookup,
		extractUser:            extractUser,
	}
}
