package topic

import (
	"context"
	"errors"
	"time"
)

type Visibility int

const (
	VisibilityPublic    Visibility = 0
	VisibilityFollowers Visibility = 1
	VisibilityPrivate   Visibility = 2
)

type Topic struct {
	ID            int
	UserID        string
	GroupID       *string
	Title         string
	Content       string
	ImagePath     string
	Visibility    Visibility
	CreatedAt     time.Time
	UpdatedAt     time.Time
	OwnerUsername string
	UpvoteCount   int
	DownvoteCount int
	VoteScore     int
	CommentsCount int
	UserVote      *int
	AllowedUsers  []string
}

type VoteCounts struct {
	Upvotes   int
	Downvotes int
	Score     int
}

var (
	ErrTopicNotFound    = errors.New("topic not found")
	ErrUnauthorized     = errors.New("user not authorized")
	ErrInvalidVoteValue = errors.New("reaction_type must be 1 (like) or -1 (dislike)")
)

type ImageStorage interface {
	Upload(ctx context.Context, data []byte, path string) error
	Delete(ctx context.Context, path string) error
}

type FollowChecker interface {
	AreConnected(ctx context.Context, followerID, followeeID string) (bool, error)
}

//nolint:interfacebloat // single repository interface for vertical slice D2 compliance
type Repository interface {
	CreateTopic(ctx context.Context, t *Topic, allowedUserIDs []string) error
	UpdateTopic(ctx context.Context, t *Topic, allowedUserIDs []string) error
	DeleteTopic(ctx context.Context, userID string, topicID int) error
	GetTopicByID(ctx context.Context, topicID int, userID *string) (*Topic, error)
	GetImagePathFromTopicID(ctx context.Context, topicID int, userID string) (string, error)

	GetFeed(ctx context.Context, userID string, page, size int, orderBy, order, filter string) ([]Topic, int, error)
	GetTopicsByUserID(ctx context.Context, ownerID, requesterID string, page, size int) ([]Topic, int, error)
	GetTopicsByGroupID(ctx context.Context, groupID string, page, size int) ([]Topic, int, error)

	CastVote(ctx context.Context, userID string, topicID int, reactionType int) error
	DeleteVote(ctx context.Context, userID string, topicID int) error
	GetVoteCounts(ctx context.Context, topicID int) (*VoteCounts, error)
	GetPostCount(ctx context.Context, userID string) (int, error)
	GetVoteCount(ctx context.Context, userID string) (int, error)
}
