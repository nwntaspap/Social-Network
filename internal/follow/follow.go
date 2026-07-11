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
	AreConnected(ctx context.Context, a, b string) (bool, error)
}
