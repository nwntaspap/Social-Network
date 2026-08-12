package bootstrap

import (
	"context"
	"net/http"

	"social-network/internal/core/middleware"
	localstorage "social-network/internal/infra/storage/local"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
	topiccommands "social-network/internal/topic/commands"
	topicqueries "social-network/internal/topic/queries"
	topicstore "social-network/internal/topic/store"
	topictransport "social-network/internal/topic/transport"
	"social-network/internal/user"
	userstore "social-network/internal/user/store"
)

func initTopic(db database.DB, bus eventbus.EventBus) *topictransport.Handler {
	store := topicstore.NewSQLiteStore(db)
	img := localstorage.NewLocalStorage()
	users := userstore.NewSQLiteStore(db)

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	userLookup := &topicUserLookupAdapter{repo: userstore.NewSQLiteStore(db)}

	return topictransport.NewHandler(
		extractUser,
		userLookup,
		topiccommands.NewCreateTopicHandler(store, img),
		topiccommands.NewUpdateTopicHandler(store, img),
		topiccommands.NewDeleteTopicHandler(store, bus, img),
		topiccommands.NewCastVoteHandler(store, bus, users),
		topiccommands.NewDeleteVoteHandler(store, bus),
		// we need to add a deleteVote Handler the command is ready we just need to wire it
		topicqueries.NewGetFeedResolver(store),
		topicqueries.NewGetTopicResolver(store),
		topicqueries.NewGetTopicsByUserResolver(store),
		topicqueries.NewGetTopicsByGroupResolver(store),
		topicqueries.NewGetVoteCountsResolver(store),
	)
}

type topicUserLookupAdapter struct {
	repo user.Repository
}

func (a *topicUserLookupAdapter) GetUserByID(ctx context.Context, id string) (*topictransport.UserResult, error) {
	u, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &topictransport.UserResult{
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

var _ topictransport.UserLookup = (*topicUserLookupAdapter)(nil)
