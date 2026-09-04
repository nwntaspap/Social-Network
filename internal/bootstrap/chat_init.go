package bootstrap

import (
	"context"
	"net/http"

	"social-network/internal/chat"
	chatcommands "social-network/internal/chat/commands"
	"social-network/internal/chat/queries"
	chatstore "social-network/internal/chat/store"
	chattransport "social-network/internal/chat/transport"
	"social-network/internal/core/middleware"
	"social-network/internal/core/realtime"
	followstore "social-network/internal/follow/store"
	"social-network/internal/platform/database"
	"social-network/internal/platform/logger"
	userstore "social-network/internal/user/store"
)

func initChat(db database.DB, hub *realtime.Hub, logger logger.Logger) *chattransport.Handler {
	store := chatstore.NewSQLiteStore(db)
	followStore := followstore.NewSQLiteStore(db)
	userStore := userstore.NewSQLiteStore(db)

	ba := &chat.BroadcasterAdapter{
		IsOnlineFn: hub.IsOnline,
	}
	ua := &chat.UserAdapter{
		GetAllFn: func(ctx context.Context) ([]*chat.UserRef, error) {
			users, err := userStore.ListAll(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]*chat.UserRef, len(users))
			for i, u := range users {
				ref := &chat.UserRef{ID: u.ID, Nickname: u.Nickname}
				ref.AvatarURL = u.AvatarPath
				result[i] = ref
			}
			return result, nil
		},
	}

	followAdapter := &chat.FollowAdapter{AreConnectedFn: followStore.AreConnected}

	getHistory := queries.NewGetChatHistoryResolver(store, followAdapter)
	getUsers := queries.NewGetChatUsersResolver(store, ua, ba)

	userLookup := &chatUserLookupAdapter{repo: userStore}

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	gate := chatcommands.NewMessageGate(followAdapter)
	startChat := chatcommands.NewOpenPrivateChatHandler(store, gate)

	return chattransport.NewHandlerWithStart(logger, extractUser, userLookup, getHistory, getUsers, startChat)
}

type chatUserLookupAdapter struct {
	repo *userstore.SQLiteStore
}

func (a *chatUserLookupAdapter) GetUserByID(ctx context.Context, id string) (*chattransport.UserResult, error) {
	u, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &chattransport.UserResult{
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

var _ chattransport.UserLookup = (*chatUserLookupAdapter)(nil)
