package queries

import (
	"context"
	"sort"
	"time"

	"social-network/internal/chat"
)

type GetChatUsersRequest struct {
	MeID string
}

type ChatUser struct {
	LastMessageAt *time.Time `json:"last_message_at"`
	UserID        string     `json:"user_id"`
	Nickname      string     `json:"nickname"`
	ChatID        string     `json:"chat_id,omitempty"`
	UnreadCount   int        `json:"unread_count"`
	IsOnline      bool       `json:"is_online"`
}

type GetChatUsersResolver struct {
	chatRepo    chat.Repository
	userRepo    chat.UserRepository
	broadcaster chat.Broadcaster
}

func NewGetChatUsersResolver(chatRepo chat.Repository, userRepo chat.UserRepository, broadcaster chat.Broadcaster) *GetChatUsersResolver {
	return &GetChatUsersResolver{
		chatRepo:    chatRepo,
		userRepo:    userRepo,
		broadcaster: broadcaster,
	}
}

func (r *GetChatUsersResolver) Resolve(ctx context.Context, req GetChatUsersRequest) ([]ChatUser, error) {
	allUsers, err := r.userRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	myChats, err := r.chatRepo.GetChatsForUser(ctx, req.MeID)
	if err != nil {
		return nil, err
	}

	lastMsgAt := make(map[string]*time.Time)
	chatIDs := make(map[string]string)
	unreadPerChat := make(map[string]int)
	for _, c := range myChats {
		otherID := c.UserTwoID
		if otherID == req.MeID {
			otherID = c.UserOneID
		}
		lastMsgAt[otherID] = c.LastMessageAt
		chatIDs[otherID] = c.ID
		unreadPerChat[otherID] = c.UnreadCount
	}

	var withMsg []ChatUser
	var withoutMsg []ChatUser

	for _, u := range allUsers {
		if u.ID == req.MeID {
			continue
		}
		cu := ChatUser{
			UserID:        u.ID,
			Nickname:      u.Nickname,
			IsOnline:      r.broadcaster.IsOnline(u.ID),
			LastMessageAt: lastMsgAt[u.ID],
		}
		if cu.LastMessageAt != nil {
			cu.ChatID = chatIDs[u.ID]
			cu.UnreadCount = unreadPerChat[u.ID]
			withMsg = append(withMsg, cu)
		} else {
			withoutMsg = append(withoutMsg, cu)
		}
	}

	sort.Slice(withoutMsg, func(i, j int) bool {
		return withoutMsg[i].Nickname < withoutMsg[j].Nickname
	})

	result := append(withMsg, withoutMsg...)
	return result, nil
}
