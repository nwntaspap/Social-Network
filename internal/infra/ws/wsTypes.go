package ws

import (
	"encoding/json"
	"time"
)

// Inbound message types (client -> server)
const (
	TypeChatSend    = "chat.send"
	TypeChatHistory = "chat.history"
	TypeMarkRead    = "chat.mark_read"
	TypePing        = "ping"
	TypeTyping      = "typing"
	TypeChatOpen    = "chat.open"
	TypeChatClose   = "chat.close"
)

// Outbound message types (server -> client)
const (
	TypeChatMessage    = "chat.message"
	TypeHistoryResult  = "chat.history_result"
	TypeError          = "error"
	TypePong           = "pong"
	TypeIsOnlineStatus = "isOnlineStatus.update"
)

// Envelope is the wrapper for every WebSocker message.
type Envelope struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

// Payloads for inbound messages
type SendPayload struct {
	ChatID          string `json:"chat_id"`
	Content         string `json:"content"`
	ClientMessageID string `json:"client_message_id,omitempty"`
}
type ChatOpenClosePayload struct {
	ChatID string `json:"chat_id"`
}

type IsOnlineStatusPayload struct {
	UserID   string `json:"user_id"`
	IsOnline bool   `json:"isOnline"`
}

type TypingPayload struct {
	UserID string `json:"user_id"`
}

type HistoryPayload struct {
	ChatID          string `json:"chat_id"`
	BeforeMessageID int    `json:"before_message_id,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type MarkReadPayload struct {
	ChatID        string `json:"chat_id"`
	UpToMessageID int    `json:"up_to_message_id"`
}

// Payloads for outbound messages
type MessagePayload struct {
	ID              int       `json:"id"`
	ChatID          string    `json:"chat_id"`
	SenderID        string    `json:"sender_id"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	ClientMessageID *string   `json:"client_message_id,omitempty"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}
