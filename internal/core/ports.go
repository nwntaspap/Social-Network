package core

import "context"

type UserPrivacyChecker interface {
	IsPrivate(ctx context.Context, userID string) (bool, error)
}

type EventBus interface {
	Publish(ctx context.Context, eventType string, payload any) error
}
