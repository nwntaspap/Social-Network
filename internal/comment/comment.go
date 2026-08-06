package comment

import (
	"context"
	"errors"
	"time"
)

type Comment struct {
	ID            int
	TopicID       int
	UserID        string
	Content       string
	ImagePath     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	UpvoteCount   int
	DownvoteCount int
	VoteScore     int
	UserVote      *int
}

type VoteCounts struct {
	Upvotes   int
	Downvotes int
	Score     int
}

var (
	ErrCommentNotFound  = errors.New("comment not found")
	ErrVoteNotFound     = errors.New("vote not found")
	ErrInvalidVoteValue = errors.New("reaction_type must be 1 (upvote) or -1 (downvote)")
)

type EventBus interface {
	Publish(ctx context.Context, eventType string, payload any) error
}

type ImageStorage interface {
	Upload(ctx context.Context, data []byte, path string) error
}

type Repository interface {
	CreateComment(ctx context.Context, c *Comment) error
	UpdateComment(ctx context.Context, c *Comment) error
	DeleteComment(ctx context.Context, userID string, commentID int) error
	GetCommentByID(ctx context.Context, commentID int) (*Comment, error)
	GetCommentByIDWithVotes(ctx context.Context, commentID int, userID *string) (*Comment, error)
	GetCommentsByTopicID(ctx context.Context, topicID int) ([]Comment, error)
	GetCommentsByTopicIDWithVotes(ctx context.Context, topicID int, userID *string) ([]Comment, error)
	CastCommentVote(ctx context.Context, userID string, commentID int, reactionType int) error
	GetVoteCounts(ctx context.Context, commentID int) (*VoteCounts, error)
	GetCommentCount(ctx context.Context, userID string) (int, error)
}
