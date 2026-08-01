package bootstrap

import (
	"context"
	"net/http"

	"social-network/internal/chat"
	"social-network/internal/chat/queries"
	chatstore "social-network/internal/chat/store"
	chattransport "social-network/internal/chat/transport"
	"social-network/internal/core/middleware"
	domainuser "social-network/internal/domain/user"
	"social-network/internal/infra/ws"
	"social-network/internal/platform/database"
	"social-network/internal/user"
	userstore "social-network/internal/user/store"
)

func initChat(db database.DB, hub *ws.Hub, userRepo domainuser.Repository) *chattransport.Handler {
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
				ref := &chat.UserRef{ID: u.ID, Nickname: u.Nickname}
				if u.AvatarURL != nil {
					ref.AvatarURL = *u.AvatarURL
				}
				result[i] = ref
			}
			return result, nil
		},
	}

	getHistory := queries.NewGetChatHistoryResolver(store)
	getUsers := queries.NewGetChatUsersResolver(store, ua, ba)

	userLookup := &chatUserLookupAdapter{repo: userstore.NewSQLiteStore(db)}

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	return chattransport.NewHandler(extractUser, userLookup, getHistory, getUsers)
}

type chatUserLookupAdapter struct {
	repo user.Repository
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
