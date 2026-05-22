package ws

import (
	"encoding/json"
)

type WSRouter interface {
	Route(client *Client, raw []byte)
}

type WSHandler interface {
	Handle(client *Client, env Envelope)
}

type wsRouter struct {
	chatHistoryHandler WSHandler
	pingHandler        WSHandler
	markAsReadHandler  WSHandler
	sendHandler        WSHandler
	chatOpenHandler    WSHandler
	chatCloseHandler   WSHandler
	chatTypingHandler  WSHandler
}

func NewWSRouter(chatHistoryHandler, pingHandler, markAsReasHandler, sendHandler, chatOpenHandler, chatCloseHandler, chatTypingHandler WSHandler) WSRouter {
	return &wsRouter{
		chatHistoryHandler: chatHistoryHandler,
		pingHandler:        pingHandler,
		markAsReadHandler:  markAsReasHandler,
		sendHandler:        sendHandler,
		chatOpenHandler:    chatOpenHandler,
		chatCloseHandler:   chatCloseHandler,
		chatTypingHandler:  chatTypingHandler,
	}
}

func (r *wsRouter) Route(client *Client, raw []byte) {
	var env Envelope
	err := json.Unmarshal(raw, &env)
	if err != nil {
		sendError(client, "", "invalid message format")
		return
	}

	switch env.Type {
	case TypePing:
		r.pingHandler.Handle(client, env)
	case TypeChatSend:
		r.sendHandler.Handle(client, env)
	case TypeChatHistory:
		r.chatHistoryHandler.Handle(client, env)
	case TypeChatOpen:
		r.chatOpenHandler.Handle(client, env)
	case TypeChatClose:
		r.chatCloseHandler.Handle(client, env)
	case TypeTyping:
		r.chatTypingHandler.Handle(client, env)
	case TypeMarkRead:
		r.markAsReadHandler.Handle(client, env)
	default:
		sendError(client, env.RequestID, "unknown message type")
	}
}

func sendError(client *Client, requestID, message string) {
	payload, _ := json.Marshal(ErrorPayload{Message: message})
	reply, _ := json.Marshal(Envelope{
		Type:      TypeError,
		RequestID: requestID,
		Payload:   payload,
	})
	client.send <- reply
}
