package bootstrap

import (
	"context"
	"net/http"

	"social-network/internal/chat"
	"social-network/internal/chat/queries"
	chatstore "social-network/internal/chat/store"
	chattransport "social-network/internal/chat/transport"
	"social-network/internal/core/middleware"
	"social-network/internal/domain/user"
	"social-network/internal/infra/ws"
	"social-network/internal/platform/database"
)

func initChat(db database.DB, hub *ws.Hub, userRepo user.Repository) *chattransport.Handler {
	store := chatstore.NewSQLiteStore(db)

	ba := &chat.BroadcasterAdapter{
		IsOnlineFn: hub.IsOnline,
	}
	ua := &chat.UserAdapter{
		GetAllFn: func(ctx context.Context) ([]*chat.UserRef, error) {
			users, err := userRepo.GetAll(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]*chat.UserRef, len(users))
			for i, u := range users {
				result[i] = &chat.UserRef{ID: u.ID, Nickname: u.Nickname}
			}
			return result, nil
		},
	}

	getHistory := queries.NewGetChatHistoryResolver(store)
	getUsers := queries.NewGetChatUsersResolver(store, ua, ba)

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	return chattransport.NewHandler(extractUser, getHistory, getUsers)
}
