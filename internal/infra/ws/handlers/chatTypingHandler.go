package handlers

import (
	"encoding/json"

	ws "github.com/arnald/forum/internal/infra/ws"
)

type ChatTypingHandler struct {
	hub *ws.Hub
}

func NewChatTypingHandler(hub *ws.Hub) *ChatTypingHandler {
	return &ChatTypingHandler{
		hub: hub,
	}
}

func (h *ChatTypingHandler) Handle(client *ws.Client, env ws.Envelope) {
	var payload ws.ChatTypingPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendError(client, env.RequestID, "invalid payload")
		return
	}
	if payload.ChatID == "" {
		sendError(client, env.RequestID, "chat_id of where user is typing is missing")
		return
	}

	outPayload, _ := json.Marshal(ws.ChatIsTyping{
		ChatID: payload.ChatID,
		UserID: client.UserID,
	})

	reply, _ := json.Marshal(ws.Envelope{
		Type:    ws.TypeIsTyping,
		Payload: outPayload,
	})

	chatObservers := h.hub.GetObserversForChat(payload.ChatID, client.UserID)
	for _, c := range chatObservers {
		c.Send(reply)
	}
}
