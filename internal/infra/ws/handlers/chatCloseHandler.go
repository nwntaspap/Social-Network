package handlers

import (
	"encoding/json"

	ws "github.com/arnald/forum/internal/infra/ws"
)

type ChatCloseHandler struct {
	hub *ws.Hub
}

func NewChatCloseHandler(hub *ws.Hub) *ChatCloseHandler {
	return &ChatCloseHandler{
		hub: hub,
	}
}

func (h *ChatCloseHandler) Handle(client *ws.Client, env ws.Envelope) {
	//to be implemented
	var payload ws.ChatOpenClosePayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendError(client, env.RequestID, "invalid payload")
		return
	}

	if payload.ChatID == "" {
		sendError(client, env.RequestID, "chat_id and content are required")
		return
	}
	h.hub.CloseChat(client)
}
