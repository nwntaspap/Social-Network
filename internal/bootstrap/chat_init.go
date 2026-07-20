package bootstrap

import (
	"context"
	"database/sql"
	"net/http"

	"social-network/internal/chat"
	"social-network/internal/chat/commands"
	"social-network/internal/chat/queries"
	chatstore "social-network/internal/chat/store"
	chattransport "social-network/internal/chat/transport"
	"social-network/internal/domain/user"
	followstore "social-network/internal/follow/store"
	"social-network/internal/infra/middleware"
	"social-network/internal/infra/ws"
)

func initChat(db *sql.DB, hub *ws.Hub, userRepo user.Repository) *chattransport.Handler {
	store := chatstore.NewSQLiteStore(db)

	followStore := followstore.NewSQLiteStore(db)

	fc := &chat.FollowAdapter{
		AreConnectedFn: followStore.AreConnected,
	}
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

	sendMsg := commands.NewSendPrivateMessageHandler(store, fc)
	markRead := commands.NewMarkAsReadHandler(store)
	getHistory := queries.NewGetChatHistoryResolver(store)
	getUsers := queries.NewGetChatUsersResolver(store, ua, ba)

	extractUser := func(r *http.Request) (string, bool) {
		user := middleware.GetUserFromContext(r)
		if user == nil {
			return "", false
		}
		return user.ID, true
	}

	return chattransport.NewHandler(extractUser, sendMsg, markRead, getHistory, getUsers)
}
