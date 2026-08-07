package commands

import (
	"context"

	"social-network/internal/chat"
)

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
	repo chat.Repository
	gate *MessageGate
}

func NewSendPrivateMessageHandler(repo chat.Repository, gate *MessageGate) *SendPrivateMessageHandler {
	return &SendPrivateMessageHandler{repo: repo, gate: gate}
}

func (h *SendPrivateMessageHandler) Execute(ctx context.Context, cmd SendPrivateMessageCommand) (SendPrivateMessageResult, error) {
	if err := h.gate.Validate(ctx, cmd.SenderID, cmd.ReceiverID); err != nil {
		return SendPrivateMessageResult{}, err
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
