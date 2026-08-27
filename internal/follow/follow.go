package follow

import (
	"context"
	"time"
)

type Follow struct {
	FollowerID string    `json:"followerId"`
	FolloweeID string    `json:"followeeId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Request struct {
	FollowerID string    `json:"followerId"`
	FolloweeID string    `json:"followeeId"`
	CreatedAt  time.Time `json:"createdAt"`
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
	GetFollowerCount(ctx context.Context, userID string) (int, error)
	GetFollowingCount(ctx context.Context, userID string) (int, error)
}

type UserPrivacyChecker interface {
	IsPrivate(ctx context.Context, userID string) (bool, error)
}

var _ UserPrivacyChecker = (*PrivacyStub)(nil)

type PrivacyStub struct{}

func (p *PrivacyStub) IsPrivate(_ context.Context, _ string) (bool, error) {
	return false, nil
}

type NoopEventBus struct{}

func (n *NoopEventBus) Publish(_ context.Context, _ string, _ any) error {
	return nil
}
