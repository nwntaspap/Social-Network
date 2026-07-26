package bootstrap

import (
	"context"
	"database/sql"
	"net/http"

	"social-network/internal/core/middleware"
	"social-network/internal/group"
	groupcommands "social-network/internal/group/commands"
	groupqueries "social-network/internal/group/queries"
	groupstore "social-network/internal/group/store"
	grouptransport "social-network/internal/group/transport"
	"social-network/internal/user"
	userstore "social-network/internal/user/store"
)

func initGroup(db *sql.DB) *grouptransport.Handler {
	store := groupstore.NewSQLiteStore(db)

	bus := &groupEventBus{}
	followChecker := &groupFollowChecker{}

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	userLookup := &userLookupAdapter{repo: userstore.NewSQLiteStore(db)}

	return grouptransport.NewHandler(
		extractUser,
		userLookup,
		groupcommands.NewCreateGroupHandler(store),
		groupcommands.NewInviteMemberHandler(store, followChecker, bus),
		groupcommands.NewRespondInviteHandler(store),
		groupcommands.NewRequestJoinHandler(store, bus),
		groupcommands.NewRespondJoinHandler(store),
		groupcommands.NewCreateGroupPostHandler(store),
		groupcommands.NewCreateGroupPostCommentHandler(store),
		groupcommands.NewLeaveGroupHandler(store),
		groupcommands.NewUpdateGroupHandler(store),
		groupcommands.NewDeleteGroupHandler(store),
		groupqueries.NewListGroupsResolver(store),
		groupqueries.NewGetGroupResolver(store),
		groupqueries.NewGetGroupFeedResolver(store),
		groupqueries.NewGetGroupChatResolver(store),
		groupqueries.NewGetGroupPostCommentsResolver(store),
		groupqueries.NewGetGroupMembersResolver(store),
	)
}

type groupEventBus struct{}

func (b *groupEventBus) Publish(_ context.Context, _ string, _ any) error {
	return nil
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
	_ group.EventBus            = (*groupEventBus)(nil)
	_ grouptransport.UserLookup = (*userLookupAdapter)(nil)
)
