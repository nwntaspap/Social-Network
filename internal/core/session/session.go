package session

import (
	"context"
	"time"
)

type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

type Manager interface {
	Create(ctx context.Context, userID string) (*Session, error)
	Get(ctx context.Context, token string) (*Session, error)
	Revoke(ctx context.Context, token string) error
}
