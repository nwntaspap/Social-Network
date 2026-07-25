package commands

import (
	"context"
	"errors"

	"social-network/internal/chat"
)

var ErrNotConnected = errors.New("users are not connected: at least one must follow the other")

type SendPrivateMessageCommand struct {
	SenderID        string
	ReceiverID      string
	Content         string
	ClientMessageID string
}

type SendPrivateMessageResult struct {
	Message     *chat.Message
	RecipientID string
	ChatID      string
}

type SendPrivateMessageHandler struct {
	repo   chat.Repository
	follow chat.FollowChecker
}

func NewSendPrivateMessageHandler(repo chat.Repository, follow chat.FollowChecker) *SendPrivateMessageHandler {
	return &SendPrivateMessageHandler{repo: repo, follow: follow}
}

func (h *SendPrivateMessageHandler) Execute(ctx context.Context, cmd SendPrivateMessageCommand) (SendPrivateMessageResult, error) {
	connected, err := h.follow.AreConnected(ctx, cmd.SenderID, cmd.ReceiverID)
	if err != nil {
		return SendPrivateMessageResult{}, err
	}
	if !connected {
		return SendPrivateMessageResult{}, ErrNotConnected
	}

	c, err := h.repo.GetOrCreateChat(ctx, cmd.SenderID, cmd.ReceiverID)
	if err != nil {
		return SendPrivateMessageResult{}, err
	}

	msg, err := h.repo.SendMessage(ctx, c.ID, cmd.SenderID, cmd.Content, cmd.ClientMessageID)
	if err != nil {
		return SendPrivateMessageResult{}, err
	}

	recipientID := c.UserOneID
	if recipientID == cmd.SenderID {
		recipientID = c.UserTwoID
	}

	return SendPrivateMessageResult{
		Message:     msg,
		RecipientID: recipientID,
		ChatID:      c.ID,
	}, nil
}
