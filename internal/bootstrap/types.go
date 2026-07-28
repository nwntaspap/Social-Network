package bootstrap

import (
	"context"

	"social-network/internal/domain/session"
)

type sessionManager interface {
	CreateSession(ctx context.Context, userID string) (*session.Session, error)
	GetSession(sessionID string) (*session.Session, error)
	DeleteSession(sessionID string) error
}
