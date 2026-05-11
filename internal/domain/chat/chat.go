package chat

import "time"

type Chat struct {
	ID            string     `json:"id"`
	UserLowID     string     `json:"user_low_id"`
	UserHighID    string     `json:"user_high_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastMessageID *int       `json:"last_message_id"`
	LastMessageAt *time.Time `json:"last_message_at"`
}
