package transport

import (
	"context"
	"encoding/json"

	"social-network/internal/chat"
	"social-network/internal/chat/commands"
	"social-network/internal/chat/queries"
	"social-network/internal/core/realtime"
)

// WSSendExecutor stores a private message via the domain command.
type WSSendExecutor interface {
	Execute(ctx context.Context, cmd commands.SendPrivateMessageCommand) (commands.SendPrivateMessageResult, error)
}

// WSMarkAsRead marks messages as read for a chat participant.
type WSMarkAsRead interface {
	MarkAsRead(ctx context.Context, chatID, userID string, upToMessageID int) error
}

// WSChatGetter looks up a chat by ID (to derive the recipient of a message).
type WSChatGetter interface {
	GetChat(ctx context.Context, chatID string) (*chat.Chat, error)
}

// WSHandler provides WebSocket handlers for private chat.
type WSHandler struct {
	hub        *realtime.Hub
	send       WSSendExecutor
	getHistory ChatHistoryResolver
	markAsRead WSMarkAsRead
	getChat    WSChatGetter
}

func NewWSHandler(
	hub *realtime.Hub,
	send WSSendExecutor,
	getHistory ChatHistoryResolver,
	markAsRead WSMarkAsRead,
	getChat WSChatGetter,
) *WSHandler {
	return &WSHandler{hub: hub, send: send, getHistory: getHistory, markAsRead: markAsRead, getChat: getChat}
}

// Handlers returns the per-type handlers to register on the realtime router.
func (h *WSHandler) Handlers() map[string]realtime.WSHandler {
	return map[string]realtime.WSHandler{
		realtime.TypePing:        realtime.HandlerFunc(h.handlePing),
		realtime.TypeChatSend:    realtime.HandlerFunc(h.handleSend),
		realtime.TypeChatHistory: realtime.HandlerFunc(h.handleHistory),
		realtime.TypeChatOpen:    realtime.HandlerFunc(h.handleOpen),
		realtime.TypeChatClose:   realtime.HandlerFunc(h.handleClose),
		realtime.TypeTyping:      realtime.HandlerFunc(h.handleTyping),
		realtime.TypeMarkRead:    realtime.HandlerFunc(h.handleMarkRead),
	}
}

func (h *WSHandler) handlePing(client *realtime.Client, env realtime.Envelope) {
	reply, _ := json.Marshal(realtime.Envelope{Type: realtime.TypePong, RequestID: env.RequestID})
	client.Send(reply)
}

func (h *WSHandler) handleSend(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.SendPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendRealtimeError(client, env.RequestID, "invalid send payload")
		return
	}

	chatID := payload.ChatID
	c, err := h.getChat.GetChat(context.Background(), chatID)
	if err != nil {
		sendRealtimeError(client, env.RequestID, err.Error())
		return
	}

	if c.UserOneID != client.UserID && c.UserTwoID != client.UserID {
		sendRealtimeError(client, env.RequestID, "you are not a participant of this chat")
		return
	}
	receiverID := c.UserTwoID
	if c.UserOneID != client.UserID {
		receiverID = c.UserOneID
	}

	result, err := h.send.Execute(context.Background(), commands.SendPrivateMessageCommand{
		SenderID:        client.UserID,
		ReceiverID:      receiverID,
		Content:         payload.Content,
		ClientMessageID: payload.ClientMessageID,
	})
	if err != nil {
		sendRealtimeError(client, env.RequestID, err.Error())
		return
	}

	outbound := toRealtimeMessage(result.Message)
	h.hub.SendToUser(result.RecipientID, env.RequestID, outbound)
	h.hub.SendToUser(client.UserID, env.RequestID, outbound)
}

func (h *WSHandler) handleHistory(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.HistoryPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendRealtimeError(client, env.RequestID, "invalid history payload")
		return
	}

	messages, err := h.getHistory.Resolve(context.Background(), queries.GetChatHistoryQuery{
		ChatID:          payload.ChatID,
		RequesterID:     client.UserID,
		BeforeMessageID: payload.BeforeMessageID,
		Limit:           payload.Limit,
	})
	if err != nil {
		sendRealtimeError(client, env.RequestID, err.Error())
		return
	}

	out, _ := json.Marshal(messages)
	reply, _ := json.Marshal(realtime.Envelope{
		Type:      realtime.TypeHistoryResult,
		RequestID: env.RequestID,
		Payload:   out,
	})
	client.Send(reply)
}

func (h *WSHandler) handleOpen(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.ChatOpenClosePayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendRealtimeError(client, env.RequestID, "invalid open payload")
		return
	}

	c, err := h.getChat.GetChat(context.Background(), payload.ChatID)
	if err != nil {
		sendRealtimeError(client, env.RequestID, err.Error())
		return
	}
	if c.UserOneID != client.UserID && c.UserTwoID != client.UserID {
		sendRealtimeError(client, env.RequestID, "not a participant of this chat")
		return
	}

	// Opening a chat marks everything received so far as read.
	if c.LastMessageID != nil {
		if err := h.markAsRead.MarkAsRead(context.Background(), payload.ChatID, client.UserID, *c.LastMessageID); err != nil {
			sendRealtimeError(client, env.RequestID, err.Error())
			return
		}
	}

	h.hub.OpenChat(client, payload.ChatID)
}

func (h *WSHandler) handleClose(client *realtime.Client, _ realtime.Envelope) {
	h.hub.CloseChat(client)
}

func (h *WSHandler) handleTyping(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.ChatTypingPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendRealtimeError(client, env.RequestID, "invalid typing payload")
		return
	}

	out, _ := json.Marshal(realtime.ChatIsTyping{ChatID: payload.ChatID, UserID: client.UserID})
	reply, _ := json.Marshal(realtime.Envelope{Type: realtime.TypeIsTyping, Payload: out})

	for _, observer := range h.hub.GetObserversForChat(payload.ChatID, client.UserID) {
		observer.Send(reply)
	}
}

func (h *WSHandler) handleMarkRead(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.MarkReadPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendRealtimeError(client, env.RequestID, "invalid mark_read payload")
		return
	}
	if err := h.markAsRead.MarkAsRead(context.Background(), payload.ChatID, client.UserID, payload.UpToMessageID); err != nil {
		sendRealtimeError(client, env.RequestID, err.Error())
	}
}

func toRealtimeMessage(m *chat.Message) *realtime.Message {
	if m == nil {
		return nil
	}
	return &realtime.Message{
		ID:              m.ID,
		ChatID:          m.ChatID,
		SenderID:        m.SenderID,
		Content:         m.Content,
		CreatedAt:       m.CreatedAt,
		ClientMessageID: m.ClientMessageID,
	}
}

func sendRealtimeError(client *realtime.Client, requestID, message string) {
	payload, _ := json.Marshal(realtime.ErrorPayload{Message: message})
	reply, _ := json.Marshal(realtime.Envelope{
		Type:      realtime.TypeError,
		RequestID: requestID,
		Payload:   payload,
	})
	client.Send(reply)
}
