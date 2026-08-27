package commands

import (
	"context"

	"social-network/internal/chat"
)

type OpenPrivateChatCommand struct {
	SenderID   string
	ReceiverID string
}

type OpenPrivateChatResult struct {
	Chat *chat.Chat
}

type OpenPrivateChatHandler struct {
	repo chat.Repository
	gate *MessageGate
}

func NewOpenPrivateChatHandler(repo chat.Repository, gate *MessageGate) *OpenPrivateChatHandler {
	return &OpenPrivateChatHandler{repo: repo, gate: gate}
}

func (h *OpenPrivateChatHandler) Execute(ctx context.Context, cmd OpenPrivateChatCommand) (OpenPrivateChatResult, error) {
	if err := h.gate.Validate(ctx, cmd.SenderID, cmd.ReceiverID); err != nil {
		return OpenPrivateChatResult{}, err
	}

	c, err := h.repo.GetOrCreateChat(ctx, cmd.SenderID, cmd.ReceiverID)
	if err != nil {
		return OpenPrivateChatResult{}, err
	}
	return OpenPrivateChatResult{Chat: c}, nil
}
