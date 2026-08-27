package transport

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"social-network/internal/comment"
	"social-network/internal/comment/commands"
	"social-network/internal/comment/queries"
	"social-network/internal/platform/logger"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type UserLookup interface {
	GetUserByID(ctx context.Context, id string) (*UserResult, error)
}

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

type DeleteCommentVoteExecutor interface {
	Execute(ctx context.Context, cmd commands.DeleteCommentVoteCommand) error
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

type UserResult struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Nickname    string `json:"nickname,omitempty"`
	AboutMe     string `json:"aboutMe,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	DateOfBirth string `json:"dateOfBirth"`
	IsPublic    bool   `json:"isPublic"`
	CreatedAt   string `json:"createdAt"`
}

func (h *Handler) lookupUser(ctx context.Context, userID string) *UserResult {
	if h.userLookup == nil || userID == "" {
		return nil
	}
	u, err := h.userLookup.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}
	return u
}

type CommentResponse struct {
	ID            string      `json:"id"`
	UserID        string      `json:"userId"`
	User          *UserResult `json:"user"`
	PostID        string      `json:"postId"`
	Content       string      `json:"content"`
	ImageURL      string      `json:"imageUrl,omitempty"`
	CreatedAt     string      `json:"createdAt"`
	UpdatedAt     string      `json:"updatedAt"`
	UpvoteCount   int         `json:"upvoteCount"`
	DownvoteCount int         `json:"downvoteCount"`
	VoteScore     int         `json:"voteScore"`
	UserVote      *int        `json:"userVote,omitempty"`
}

type VoteCountsResponse struct {
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`
	Score     int `json:"score"`
}

func toCommentResponse(c *comment.Comment, user *UserResult) CommentResponse {
	return CommentResponse{
		ID:            strconv.Itoa(c.ID),
		UserID:        c.UserID,
		User:          user,
		PostID:        strconv.Itoa(c.TopicID),
		Content:       c.Content,
		ImageURL:      c.ImagePath,
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     c.UpdatedAt.Format(time.RFC3339),
		UpvoteCount:   c.UpvoteCount,
		DownvoteCount: c.DownvoteCount,
		VoteScore:     c.VoteScore,
		UserVote:      c.UserVote,
	}
}

type Handler struct {
	createComment     CreateCommentExecutor
	updateComment     UpdateCommentExecutor
	deleteComment     DeleteCommentExecutor
	castCommentVote   CastCommentVoteExecutor
	deleteCommentVote DeleteCommentVoteExecutor
	getComment        GetCommentByIDResolver
	getCommentWV      GetCommentByIDWithVotesResolver
	getByTopic        GetCommentsByTopicResolver
	getByTopicWV      GetCommentsByTopicWithVotesResolver
	getCommentVotes   GetVoteCountsResolver
	userLookup        UserLookup
	extractUser       UserExtractor
	logger            logger.Logger
}

func NewHandler(
	extractUser UserExtractor,
	userLookup UserLookup,
	createComment CreateCommentExecutor,
	updateComment UpdateCommentExecutor,
	deleteComment DeleteCommentExecutor,
	castCommentVote CastCommentVoteExecutor,
	deleteCommentVote DeleteCommentVoteExecutor,
	getComment GetCommentByIDResolver,
	getCommentWV GetCommentByIDWithVotesResolver,
	getByTopic GetCommentsByTopicResolver,
	getByTopicWV GetCommentsByTopicWithVotesResolver,
	getCommentVotes GetVoteCountsResolver,
	logger logger.Logger,
) *Handler {
	return &Handler{
		createComment:     createComment,
		updateComment:     updateComment,
		deleteComment:     deleteComment,
		castCommentVote:   castCommentVote,
		deleteCommentVote: deleteCommentVote,
		getComment:        getComment,
		getCommentWV:      getCommentWV,
		getByTopic:        getByTopic,
		getByTopicWV:      getByTopicWV,
		getCommentVotes:   getCommentVotes,
		extractUser:       extractUser,
		userLookup:        userLookup,
		logger:            logger,
	}
}
