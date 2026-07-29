package bootstrap

import (
	"context"
	"net/http"

	"social-network/internal/core/middleware"
	"social-network/internal/group"
	groupcommands "social-network/internal/group/commands"
	groupqueries "social-network/internal/group/queries"
	groupstore "social-network/internal/group/store"
	grouptransport "social-network/internal/group/transport"
	localstorage "social-network/internal/infra/storage/local"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
	"social-network/internal/user"
	userstore "social-network/internal/user/store"
)

func initGroup(db database.DB, isOnline func(string) bool) *grouptransport.Handler {
func initGroup(db database.DB, bus eventbus.EventBus, isOnline func(string) bool) *grouptransport.Handler {
	store := groupstore.NewSQLiteStore(db)
	img := localstorage.NewLocalStorage()
	users := userstore.NewSQLiteStore(db)

	followChecker := &groupFollowChecker{}

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	return grouptransport.NewHandler(
		extractUser,
		userLookup,
		groupcommands.NewCreateGroupHandler(store),
		groupcommands.NewInviteMemberHandler(store, followChecker, bus, users),
		groupcommands.NewRespondInviteHandler(store),
		groupcommands.NewRequestJoinHandler(store, bus, users),
		groupcommands.NewRespondJoinHandler(store),
		groupcommands.NewCreateGroupPostHandler(store, img),
		groupcommands.NewCreateGroupPostCommentHandler(store, img),
		groupcommands.NewCastGroupPostVoteHandler(store),
		groupcommands.NewLeaveGroupHandler(store),
		groupcommands.NewUpdateGroupHandler(store),
		groupcommands.NewDeleteGroupHandler(store),
		groupqueries.NewListGroupsResolver(store),
		groupqueries.NewGetGroupResolver(store),
		groupqueries.NewGetGroupFeedResolver(store),
		groupqueries.NewGetGroupChatResolver(store),
		groupqueries.NewGetGroupPostCommentsResolver(store),
		groupqueries.NewGetGroupMembersResolver(store),
		groupqueries.NewGetPendingInvitationsResolver(store),
		groupqueries.NewGetPendingJoinRequestsResolver(store),
		groupqueries.NewListMyGroupsResolver(store),
		groupqueries.NewGetGroupPresenceResolver(store, isOnline),
	)
}

type groupFollowChecker struct{}

func (fc *groupFollowChecker) AreConnected(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

type userLookupAdapter struct {
	repo user.Repository
}

func (a *userLookupAdapter) GetUserByID(ctx context.Context, id string) (*grouptransport.UserResult, error) {
	u, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &grouptransport.UserResult{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Nickname,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		AboutMe:   u.AboutMe,
		IsPublic:  !u.IsPrivate,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if !u.DateOfBirth.IsZero() {
		result.DateOfBirth = u.DateOfBirth.Format("2006-01-02")
	}

	if u.AvatarPath != "" {
		result.AvatarURL = u.AvatarPath
	}

	return result, nil
}

var (
	_ group.FollowChecker       = (*groupFollowChecker)(nil)
	_ grouptransport.UserLookup = (*userLookupAdapter)(nil)
)
