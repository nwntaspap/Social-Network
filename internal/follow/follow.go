package follow

import (
	"context"
	"time"
)

type Follow struct {
	FollowerID string
	FolloweeID string
	CreatedAt  time.Time
}

type Request struct {
	FollowerID string
	FolloweeID string
	CreatedAt  time.Time
}

type Repository interface {
	CreateFollow(ctx context.Context, f *Follow) error
	DeleteFollow(ctx context.Context, followerID, followeeID string) error
	GetFollowers(ctx context.Context, userID string) ([]Follow, error)
	GetFollowing(ctx context.Context, userID string) ([]Follow, error)
	CreateFollowRequest(ctx context.Context, req *Request) error
	DeleteFollowRequest(ctx context.Context, followerID, followeeID string) error
	GetPendingRequests(ctx context.Context, userID string) ([]Request, error)
	AreConnected(ctx context.Context, a, b string) (bool, error)
}

type UserPrivacyChecker interface {
	IsPrivate(ctx context.Context, userID string) (bool, error)
}

type EventBus interface {
	Publish(ctx context.Context, eventType string, payload any) error
}

var _ UserPrivacyChecker = (*PrivacyStub)(nil)

type PrivacyStub struct{}

func (p *PrivacyStub) IsPrivate(_ context.Context, _ string) (bool, error) {
	return false, nil
}

var _ EventBus = (*NoopEventBus)(nil)

type NoopEventBus struct{}

func (n *NoopEventBus) Publish(_ context.Context, _ string, _ any) error {
	return nil
}
