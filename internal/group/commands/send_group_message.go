package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"social-network/internal/group"
	"social-network/internal/pkg/uuid"
)

type SendGroupMessageCommand struct {
	GroupID  string
	SenderID string
	Content  string
}

type SendGroupMessageResult struct {
	GroupID string
	Message *group.ChatMessage
}

type SendGroupMessageHandler struct {
	repo group.Repository
}

func NewSendGroupMessageHandler(repo group.Repository) *SendGroupMessageHandler {
	return &SendGroupMessageHandler{repo: repo}
}

func (h *SendGroupMessageHandler) Execute(ctx context.Context, cmd SendGroupMessageCommand) (SendGroupMessageResult, error) {
	if cmd.GroupID == "" || cmd.SenderID == "" {
		return SendGroupMessageResult{}, errors.New("group_id and sender_id are required")
	}
	content := strings.TrimSpace(cmd.Content)
	if content == "" {
		return SendGroupMessageResult{}, errors.New("content cannot be empty")
	}

	isMember, err := h.repo.IsMember(ctx, cmd.GroupID, cmd.SenderID)
	if err != nil {
		return SendGroupMessageResult{}, err
	}
	if !isMember {
		return SendGroupMessageResult{}, group.ErrNotMember
	}

	msg := &group.ChatMessage{
		ID:        uuid.NewProvider().NewUUID(),
		GroupID:   cmd.GroupID,
		SenderID:  cmd.SenderID,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
	if err := h.repo.SendGroupChatMessage(ctx, msg); err != nil {
		return SendGroupMessageResult{}, err
	}

	return SendGroupMessageResult{GroupID: cmd.GroupID, Message: msg}, nil
}
