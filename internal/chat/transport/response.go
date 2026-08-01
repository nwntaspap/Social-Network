package transport

import (
	"strconv"
	"time"

	"social-network/internal/chat"
	"social-network/internal/chat/queries"
)

type ChatUserResponse struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	AvatarURL     string `json:"avatarUrl,omitempty"`
	IsOnline      bool   `json:"isOnline"`
	LastMessageAt string `json:"lastMessageAt,omitempty"`
}

type ConversationResponse struct {
	ID           string             `json:"id"`
	Participants []ChatUserResponse `json:"participants"`
	UnreadCount  int                `json:"unreadCount"`
	CreatedAt    string             `json:"createdAt"`
}

func toConversationResponse(c queries.Conversation) ConversationResponse {
	lastMessageAt := ""
	if c.LastMessageAt != nil {
		lastMessageAt = c.LastMessageAt.Format(time.RFC3339)
	}

	participant := ChatUserResponse{
		ID:            c.OtherUser.ID,
		Username:      c.OtherUser.Nickname,
		AvatarURL:     c.OtherUser.AvatarURL,
		IsOnline:      c.IsOnline,
		LastMessageAt: lastMessageAt,
	}

	return ConversationResponse{
		ID:           c.ID,
		Participants: []ChatUserResponse{participant},
		UnreadCount:  c.UnreadCount,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
	}
}

type ChatMessageResponse struct {
	ID        string      `json:"id"`
	SenderID  string      `json:"senderId"`
	Sender    *UserResult `json:"sender"`
	Content   string      `json:"content"`
	Type      string      `json:"type"`
	CreatedAt string      `json:"createdAt"`
}

func toMessageResponse(m *chat.Message, sender *UserResult) ChatMessageResponse {
	return ChatMessageResponse{
		ID:        strconv.Itoa(m.ID),
		SenderID:  m.SenderID,
		Sender:    sender,
		Content:   m.Content,
		Type:      "private",
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}
