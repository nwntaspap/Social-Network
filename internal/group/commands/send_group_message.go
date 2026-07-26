package commands

import (
	"context"
	"errors"

	"social-network/internal/group"
)

type SendGroupMessageCommand struct {
	GroupID   string
	SenderID  string
	Content   string
	RequestID string
}

type SendGroupMessageResult struct {
	GroupID string
}

type SendGroupMessageHandler struct {
	repo group.Repository
}

func NewSendGroupMessageHandler(repo group.Repository) *SendGroupMessageHandler {
	return &SendGroupMessageHandler{repo: repo}
}

func (h *SendGroupMessageHandler) Execute(ctx context.Context, cmd SendGroupMessageCommand) (SendGroupMessageResult, error) {
	if cmd.GroupID == "" || cmd.SenderID == "" || cmd.Content == "" {
		return SendGroupMessageResult{}, errors.New("group_id, sender_id, and content are required")
	}

	isMember, err := h.repo.IsMember(ctx, cmd.GroupID, cmd.SenderID)
	if err != nil {
		return SendGroupMessageResult{}, err
	}
	if !isMember {
		return SendGroupMessageResult{}, group.ErrNotMember
	}

	return SendGroupMessageResult{GroupID: cmd.GroupID}, nil
}
