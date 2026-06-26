package handlers

import (
	"context"
	"encoding/json"

	"social-network/internal/infra/logger"

	chatcommands "social-network/internal/app/chat/commands"

	ws "social-network/internal/infra/ws"
)

type ChatMarkReadHandler struct {
	markAsRead chatcommands.MarkAsReadHandler
	logger     logger.Logger
}

func NewChatMarkReadHandler(markAsReadHandler chatcommands.MarkAsReadHandler, logger logger.Logger) *ChatMarkReadHandler {
	return &ChatMarkReadHandler{
		markAsRead: markAsReadHandler,
		logger:     logger,
	}
}

func (h *ChatMarkReadHandler) Handle(client *ws.Client, env ws.Envelope) {
	var payload ws.MarkReadPayload
	err := json.Unmarshal(env.Payload, &payload)
	if err != nil {
		h.logger.PrintError(err, nil)
		sendError(client, env.RequestID, "invalid payload")
		return
	}

	err = h.markAsRead.Handle(context.Background(), chatcommands.MarkAsReadRequest{
		ChatID:        payload.ChatID,
		UserID:        client.UserID,
		UpToMessageID: payload.UpToMessageID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		sendError(client, env.RequestID, "failed to mark as read")
	}
}
