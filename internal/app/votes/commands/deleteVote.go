package votecommands

import (
	"context"

	"social-network/internal/domain/vote"
)

type DeleteVoteRequest struct {
	TopicID   *int
	CommentID *int
	UserID    string
}

type deleteVoteRequestHandler struct {
	VoteRepo vote.Repository
}

type DeleteVoteRequestHandler interface {
	Handle(ctx context.Context, req DeleteVoteRequest) error
}

func NewDeleteVoteHandler(repo vote.Repository) DeleteVoteRequestHandler {
	return &deleteVoteRequestHandler{
		VoteRepo: repo,
	}
}

func (h *deleteVoteRequestHandler) Handle(ctx context.Context, req DeleteVoteRequest) error {
	err := h.VoteRepo.DeleteVote(ctx, req.UserID, req.TopicID, req.CommentID)
	if err != nil {
		return err
	}

	return nil
}
