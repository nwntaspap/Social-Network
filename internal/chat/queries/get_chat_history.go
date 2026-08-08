package queries

import (
	"context"

	"social-network/internal/chat"
)

type GetChatHistoryQuery struct {
	ChatID          string
	RequesterID     string
	BeforeMessageID int
	Limit           int
}

type GetChatHistoryResolver struct {
	repo   chat.Repository
	follow chat.FollowChecker
}

func NewGetChatHistoryResolver(repo chat.Repository, follow chat.FollowChecker) *GetChatHistoryResolver {
	return &GetChatHistoryResolver{repo: repo, follow: follow}
}

func (r *GetChatHistoryResolver) Resolve(ctx context.Context, q GetChatHistoryQuery) ([]*chat.Message, error) {
	if q.RequesterID != "" {
		if err := r.validateAccess(ctx, q); err != nil {
			return nil, err
		}
	}

	limit := q.Limit
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	if q.BeforeMessageID > 0 {
		return r.repo.GetMessagesForChatBefore(ctx, q.ChatID, q.BeforeMessageID, limit)
	}
	return r.repo.GetMessagesForChat(ctx, q.ChatID, limit)
}

// validateAccess verifies that the requester is a participant of the chat and
// that the two users are still connected (at least one follows the other).
func (r *GetChatHistoryResolver) validateAccess(ctx context.Context, q GetChatHistoryQuery) error {
	c, err := r.repo.GetChat(ctx, q.ChatID)
	if err != nil {
		return err
	}
	if c.UserOneID != q.RequesterID && c.UserTwoID != q.RequesterID {
		return chat.ErrNotParticipant
	}

	other := c.UserOneID
	if c.UserOneID == q.RequesterID {
		other = c.UserTwoID
	}
	requesterFollows, err := r.follow.AreConnected(ctx, q.RequesterID, other)
	if err != nil {
		return err
	}
	otherFollows, err := r.follow.AreConnected(ctx, other, q.RequesterID)
	if err != nil {
		return err
	}
	if !requesterFollows && !otherFollows {
		return chat.ErrNotConnected
	}
	return nil
}
