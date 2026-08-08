package chat

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNotParticipant is returned when a user requests history for a chat
	// they are not part of.
	ErrNotParticipant = errors.New("user is not a participant of this chat")

	// ErrChatNotFound is returned when a chat does not exist.
	ErrChatNotFound = errors.New("chat not found")

	// ErrNotConnected is returned when the two users in a chat no longer
	// follow each other and therefore cannot exchange messages.
	ErrNotConnected = errors.New("users are not connected: at least one must follow the other")
)

type Chat struct {
	ID            string     `json:"id"`
	UserOneID     string     `json:"user_one_id"`
	UserTwoID     string     `json:"user_two_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastMessageID *int       `json:"last_message_id,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	UnreadCount   int        `json:"unread_count"`
}

type Message struct {
	ID              int       `json:"id"`
	ChatID          string    `json:"chat_id"`
	SenderID        string    `json:"sender_id"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	ClientMessageID *string   `json:"client_message_id,omitempty"`
}

// Repository is implemented by store/sqlite.go.
type Repository interface {
	// Used by: commands/send_private_msg.go
	GetOrCreateChat(ctx context.Context, userOneID, userTwoID string) (*Chat, error)

	// Used by: commands/send_private_msg.go, transport/ws.go
	GetChat(ctx context.Context, chatID string) (*Chat, error)

	// Used by: queries/get_chat_users.go
	GetChatsForUser(ctx context.Context, userID string) ([]*Chat, error)

	// Used by: commands/send_private_msg.go
	SendMessage(ctx context.Context, chatID, senderID, content, clientMessageID string) (*Message, error)

	// Used by: queries/get_chat_history.go
	GetMessagesForChat(ctx context.Context, chatID string, limit int) ([]*Message, error)

	// Used by: queries/get_chat_history.go
	GetMessagesForChatBefore(ctx context.Context, chatID string, beforeID int, limit int) ([]*Message, error)

	// Used by: commands/mark_as_read.go
	MarkAsRead(ctx context.Context, chatID, userID string, upToMessageID int) error

	// Used by: queries/get_chat_users.go
	GetUnreadCount(ctx context.Context, chatID, userID string) (int, error)

	// Used by: queries/get_chat_users.go
	GetAllUnreadCounts(ctx context.Context, userID string) (map[string]int, error)
}

// FollowChecker is a cross-slice interface: defined locally here,
// satisfied by follow.Repository, wired in bootstrap.
type FollowChecker interface {
	AreConnected(ctx context.Context, a, b string) (bool, error)
}

// Broadcaster checks online status (used by getChatUsers).
type Broadcaster interface {
	IsOnline(userID string) bool
}

// UserRepository is a local interface for user lookups (avoids importing domain/user).
type UserRepository interface {
	GetAll(ctx context.Context) ([]*UserRef, error)
}

type UserRef struct {
	ID        string
	Nickname  string
	AvatarURL string
}

// FollowAdapter wraps a function to satisfy FollowChecker.
// Defined here (in chat domain) so bootstrap doesn't create types satisfying chat interfaces.
type FollowAdapter struct {
	AreConnectedFn func(ctx context.Context, a, b string) (bool, error)
}

func (fa *FollowAdapter) AreConnected(ctx context.Context, x, y string) (bool, error) {
	return fa.AreConnectedFn(ctx, x, y)
}

// BroadcasterAdapter wraps a function to satisfy Broadcaster.
type BroadcasterAdapter struct {
	IsOnlineFn func(userID string) bool
}

func (a *BroadcasterAdapter) IsOnline(userID string) bool {
	return a.IsOnlineFn(userID)
}

// UserPrivacyChecker reports whether a user's profile is private.
// Satisfied by the user store in bootstrap.
type UserPrivacyChecker interface {
	IsPrivate(ctx context.Context, userID string) (bool, error)
}

// PrivacyAdapter wraps a function to satisfy UserPrivacyChecker.
type PrivacyAdapter struct {
	IsPrivateFn func(ctx context.Context, userID string) (bool, error)
}

func (a *PrivacyAdapter) IsPrivate(ctx context.Context, userID string) (bool, error) {
	return a.IsPrivateFn(ctx, userID)
}

// UserAdapter wraps a function to satisfy UserRepository.
type UserAdapter struct {
	GetAllFn func(ctx context.Context) ([]*UserRef, error)
}

func (a *UserAdapter) GetAll(ctx context.Context) ([]*UserRef, error) {
	return a.GetAllFn(ctx)
}
