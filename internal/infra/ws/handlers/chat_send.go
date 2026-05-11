package handlers

import (
	"context"
	"encoding/json"
	chatcommands "github.com/arnald/forum/internal/app/chat/commands"
	"github.com/arnald/forum/internal/infra/logger"
	ws "github.com/arnald/forum/internal/infra/ws"
)

type ChatSendHandler struct {
	sendChat chatcommands.SendChatHandler
	logger   logger.Logger
}

func NewChatSendHandler(sendChatHandler chatcommands.SendChatHandler, logger logger.Logger) *ChatSendHandler {
	return &ChatSendHandler{
		sendChat: sendChatHandler,
		logger:   logger,
	}
}

func (h *ChatSendHandler) Handle(client *ws.Client, env ws.Envelope) {
	var payload ws.SendPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendError(client, env.RequestID, "invalid payload")
		return
	}
	if payload.ChatID == "" || payload.Content == "" {
		sendError(client, env.RequestID, "chat_id and content are required")
		return
	}

	ctx := context.Background()
	res, err := h.sendChat.Handle(ctx, chatcommands.SendChatRequest{
		ChatID:          payload.ChatID,
		SenderID:        client.UserID,
		Content:         payload.Content,
		ClientMessageID: payload.ClientMessageID,
		RequestID:       env.RequestID,
	})
	if err != nil {
		sendError(client, env.RequestID, "failed to send message")
		h.logger.PrintError(err, nil)
		return
	}

	outPayload, _ := json.Marshal(ws.MessagePayload{
		ID:              res.Msg.ID,
		ChatID:          res.Msg.ChatID,
		SenderID:        res.Msg.SenderID,
		Content:         res.Msg.Content,
		CreatedAt:       res.Msg.CreatedAt,
		ClientMessageID: res.Msg.ClientMessageID,
	})
	reply, _ := json.Marshal(ws.Envelope{
		Type:      ws.TypeChatMessage,
		RequestID: env.RequestID,
		Payload:   outPayload,
	})

	//this is considered the response in the websocket level
	client.Send(reply)
	//i cannot decide whether this broadcast should be part of the ws handler or the app use-case
	// h.hub.Send(res.RecipientID, reply)
}
