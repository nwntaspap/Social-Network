package bootstrap

import (
	"context"
	"net/http"

	"social-network/internal/core/middleware"
	"social-network/internal/follow"
	followcommands "social-network/internal/follow/commands"
	followqueries "social-network/internal/follow/queries"
	followstore "social-network/internal/follow/store"
	followtransport "social-network/internal/follow/transport"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
	userstore "social-network/internal/user/store"
)

func initFollow(db database.DB, bus eventbus.EventBus) *followtransport.Handler {
	store := followstore.NewSQLiteStore(db)
	userStore := userstore.NewSQLiteStore(db)
	privacy := &followPrivacyAdapter{repo: userStore}
	bus := &follow.NoopEventBus{}
	userLookup := &followUserLookupAdapter{repo: userStore}

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	return followtransport.NewHandler(
		extractUser,
		followcommands.NewFollowUserHandler(store, privacy, bus, userStore),
		followcommands.NewUnfollowUserHandler(store, bus, userStore),
		followcommands.NewAcceptRequestHandler(store, bus, userStore),
		followcommands.NewDeclineRequestHandler(store, bus, userStore),
		followqueries.NewGetFollowersResolver(store),
		followqueries.NewGetFollowingResolver(store),
		followqueries.NewGetPendingRequestsResolver(store),
		followqueries.NewAreConnectedResolver(store),
	)
}

type followUserLookupAdapter struct {
	repo user.Repository
}

func (a *followUserLookupAdapter) GetUserByID(ctx context.Context, id string) (*followtransport.UserResult, error) {
	u, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &followtransport.UserResult{
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

var _ followtransport.UserLookup = (*followUserLookupAdapter)(nil)

type followPrivacyAdapter struct {
	repo user.Repository
}

func (a *followPrivacyAdapter) IsPrivate(ctx context.Context, userID string) (bool, error) {
	u, err := a.repo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.IsPrivate, nil
}

var _ follow.UserPrivacyChecker = (*followPrivacyAdapter)(nil)
