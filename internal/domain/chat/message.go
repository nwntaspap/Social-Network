package chat

import "time"

type Message struct {
	ID              int        `json:"id"`
	ChatID          string     `json:"chat_id"`
	SenderID        string     `json:"sender_id"`
	Content         string     `json:"content"`
	CreatedAt       time.Time  `json:"created_at"`
	ClientMessageID *string    `json:"client_message_id,omitempty"`
}
