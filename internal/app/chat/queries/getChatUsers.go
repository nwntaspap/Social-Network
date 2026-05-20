package chatqueries

import (
	"context"
	"sort"
	"time"

	chatapp "github.com/arnald/forum/internal/app/chat"
	"github.com/arnald/forum/internal/domain/chat"
	"github.com/arnald/forum/internal/domain/user"
)

type GetChatUsersRequest struct {
	MeID string
}

type ChatUser struct {
	UserID        string     `json:"user_id"`
	Nickname      string     `json:"nickname"`
	IsOnline      bool       `json:"is_online"`
	LastMessageAt *time.Time `json:"last_message_at"`
	ChatId        string     `json:"chat_id,omitempty"`
	UnreadCount   int        `json:"unread_count"`
}
type GetChatUsersHandler interface {
	Handle(ctx context.Context, req GetChatUsersRequest) ([]ChatUser, error)
}

type getChatUsersHandler struct {
	chatRepo    chat.Repository
	userRepo    user.Repository
	broadcaster chatapp.Broadcaster
}

func NewGetChatUsersHandler(chatRepo chat.Repository, userRepo user.Repository, broadCaster chatapp.Broadcaster) GetChatUsersHandler {
	return &getChatUsersHandler{
		chatRepo:    chatRepo,
		userRepo:    userRepo,
		broadcaster: broadCaster,
	}
}

func (h *getChatUsersHandler) Handle(ctx context.Context, req GetChatUsersRequest) ([]ChatUser, error) {
	allUsers, err := h.userRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	myChats, err := h.chatRepo.GetChatsForUser(ctx, req.MeID)
	if err != nil {
		return nil, err
	}

	lastMsgAt := make(map[string]*time.Time)
	users_chatId := make(map[string]string)
	unreadCountPerUserChat := make(map[string]int)
	for _, c := range myChats {
		otherID := c.UserHighID
		if otherID == req.MeID {
			otherID = c.UserLowID
		}
		lastMsgAt[otherID] = c.LastMessageAt
		users_chatId[otherID] = c.ID
		unreadCountPerUserChat[otherID] = c.UnreadCount
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
			IsOnline:      h.broadcaster.IsOnline(u.ID),
			LastMessageAt: lastMsgAt[u.ID],
		}
		if cu.LastMessageAt != nil {
			cu.ChatId = users_chatId[u.ID]
			cu.UnreadCount = unreadCountPerUserChat[u.ID]
			withMsg = append(withMsg, cu)
		} else {
			withoutMsg = append(withoutMsg, cu)
		}
	}

	// withMsg: already sorted by last_message_at DESC from GetChatsForUser
	// withoutMsg: sort alphabetically by nickname
	sort.Slice(withoutMsg, func(i, j int) bool {
		return withoutMsg[i].Nickname < withoutMsg[j].Nickname
	})

	result := append(withMsg, withoutMsg...)
	return result, nil
}
