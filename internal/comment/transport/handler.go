package transport

import (
	"context"
	"net/http"
	"time"

	"social-network/internal/comment"
	"social-network/internal/comment/commands"
	"social-network/internal/comment/queries"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type CreateCommentExecutor interface {
	Execute(ctx context.Context, cmd commands.CreateCommentCommand) (*comment.Comment, error)
}

type UpdateCommentExecutor interface {
	Execute(ctx context.Context, cmd commands.UpdateCommentCommand) error
}

type DeleteCommentExecutor interface {
	Execute(ctx context.Context, cmd commands.DeleteCommentCommand) error
}

type CastCommentVoteExecutor interface {
	Execute(ctx context.Context, cmd commands.CastCommentVoteCommand) error
}

type GetCommentByIDResolver interface {
	Resolve(ctx context.Context, q queries.GetCommentByIDQuery) (*comment.Comment, error)
}

type GetCommentByIDWithVotesResolver interface {
	Resolve(ctx context.Context, q queries.GetCommentByIDWithVotesQuery) (*comment.Comment, error)
}

type GetCommentsByTopicResolver interface {
	Resolve(ctx context.Context, q queries.GetCommentsByTopicQuery) ([]comment.Comment, error)
}

type GetCommentsByTopicWithVotesResolver interface {
	Resolve(ctx context.Context, q queries.GetCommentsByTopicWithVotesQuery) ([]comment.Comment, error)
}

type GetVoteCountsResolver interface {
	Resolve(ctx context.Context, q queries.GetVoteCountsQuery) (*comment.VoteCounts, error)
}

type CommentResponse struct {
	ID            int    `json:"id"`
	UserID        string `json:"userId"`
	TopicID       int    `json:"topicId"`
	Content       string `json:"content"`
	ImagePath     string `json:"imagePath,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
	UpvoteCount   int    `json:"upvoteCount"`
	DownvoteCount int    `json:"downvoteCount"`
	VoteScore     int    `json:"voteScore"`
	UserVote      *int   `json:"userVote,omitempty"`
}

type VoteCountsResponse struct {
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`
	Score     int `json:"score"`
}

func toCommentResponse(c *comment.Comment) CommentResponse {
	return CommentResponse{
		ID:            c.ID,
		UserID:        c.UserID,
		TopicID:       c.TopicID,
		Content:       c.Content,
		ImagePath:     c.ImagePath,
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     c.UpdatedAt.Format(time.RFC3339),
		UpvoteCount:   c.UpvoteCount,
		DownvoteCount: c.DownvoteCount,
		VoteScore:     c.VoteScore,
		UserVote:      c.UserVote,
	}
}

type Handler struct {
	createComment   CreateCommentExecutor
	updateComment   UpdateCommentExecutor
	deleteComment   DeleteCommentExecutor
	castCommentVote CastCommentVoteExecutor
	getComment      GetCommentByIDResolver
	getCommentWV    GetCommentByIDWithVotesResolver
	getByTopic      GetCommentsByTopicResolver
	getByTopicWV    GetCommentsByTopicWithVotesResolver
	getCommentVotes GetVoteCountsResolver
	extractUser     UserExtractor
}

func NewHandler(
	extractUser UserExtractor,
	createComment CreateCommentExecutor,
	updateComment UpdateCommentExecutor,
	deleteComment DeleteCommentExecutor,
	castCommentVote CastCommentVoteExecutor,
	getComment GetCommentByIDResolver,
	getCommentWV GetCommentByIDWithVotesResolver,
	getByTopic GetCommentsByTopicResolver,
	getByTopicWV GetCommentsByTopicWithVotesResolver,
	getCommentVotes GetVoteCountsResolver,
) *Handler {
	return &Handler{
		createComment:   createComment,
		updateComment:   updateComment,
		deleteComment:   deleteComment,
		castCommentVote: castCommentVote,
		getComment:      getComment,
		getCommentWV:    getCommentWV,
		getByTopic:      getByTopic,
		getByTopicWV:    getByTopicWV,
		getCommentVotes: getCommentVotes,
		extractUser:     extractUser,
	}
}
