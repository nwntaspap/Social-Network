package handlers

import (
	"context"
	"encoding/json"

	chatqueries "github.com/arnald/forum/internal/app/chat/queries"
	"github.com/arnald/forum/internal/domain/chat"
	"github.com/arnald/forum/internal/infra/logger"
	"github.com/arnald/forum/internal/infra/ws"
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
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
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
