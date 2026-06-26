package handlers

import (
	"context"
	"encoding/json"

	"social-network/internal/domain/chat"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/ws"

	chatqueries "social-network/internal/app/chat/queries"
)

type ChatHistoryHandler struct {
	getChatHistory chatqueries.GetChatHistoryHandler
	logger         logger.Logger
}

func NewChatHistoryHandler(getChatHistoryHandler chatqueries.GetChatHistoryHandler, logger logger.Logger) *ChatHistoryHandler {
	return &ChatHistoryHandler{
		getChatHistory: getChatHistoryHandler,
		logger:         logger,
	}
}

func (h *ChatHistoryHandler) Handle(client *ws.Client, env ws.Envelope) {
	var payload ws.HistoryPayload
	err := json.Unmarshal(env.Payload, &payload)
	if err != nil {
		sendError(client, env.RequestID, "invalid payload")
		return
	}

	if payload.ChatID == "" {
		sendError(client, env.RequestID, "chat_id is required")
		return
	}

	ctx := context.Background()

	messages, err := h.getChatHistory.Handle(ctx, chatqueries.GetChatHistoryRequest{
		ChatID:          payload.ChatID,
		BeforeMessageID: payload.BeforeMessageID,
		Limit:           payload.Limit,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		sendError(client, env.RequestID, "failed to load messages")
		return
	}

	// Handle nil messages and return empty array instead
	if messages == nil {
		messages = make([]*chat.Message, 0)
	}

	outPayload, _ := json.Marshal(messages)
	reply, _ := json.Marshal(ws.Envelope{
		Type:      ws.TypeHistoryResult,
		RequestID: env.RequestID,
		Payload:   outPayload,
	})
	client.Send(reply)
}
