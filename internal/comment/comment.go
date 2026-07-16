package comment

import (
	"context"
	"errors"
	"time"
)

type Comment struct {
	ID        int
	TopicID   int
	UserID    string
	Content   string
	ImagePath string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var ErrCommentNotFound = errors.New("comment not found")

type Repository interface {
	CreateComment(ctx context.Context, c *Comment) error
	UpdateComment(ctx context.Context, c *Comment) error
	DeleteComment(ctx context.Context, userID string, commentID int) error
	GetCommentByID(ctx context.Context, commentID int) (*Comment, error)
	GetCommentByIDWithVotes(ctx context.Context, commentID int, userID *string) (*Comment, error)
	GetCommentsByTopicID(ctx context.Context, topicID int) ([]Comment, error)
	GetCommentsByTopicIDWithVotes(ctx context.Context, topicID int, userID *string) ([]Comment, error)
}
