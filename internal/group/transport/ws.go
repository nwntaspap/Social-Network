package transport

import (
	"context"
	"encoding/json"

	"social-network/internal/core/realtime"
	"social-network/internal/group/commands"
	"social-network/internal/group/queries"
)

// GroupMemberIDsResolver lists the members of a group.
type GroupMemberIDsResolver interface {
	Resolve(ctx context.Context, q queries.ListGroupMemberIDsQuery) ([]string, error)
}

// GroupWSHandler provides WebSocket handlers for group chat.
type GroupWSHandler struct {
	hub        *realtime.Hub
	send       SendGroupMessageExecutor
	getHistory GetGroupChatResolver
	memberIDs  GroupMemberIDsResolver
}

func NewGroupWSHandler(
	hub *realtime.Hub,
	send SendGroupMessageExecutor,
	getHistory GetGroupChatResolver,
	memberIDs GroupMemberIDsResolver,
) *GroupWSHandler {
	return &GroupWSHandler{hub: hub, send: send, getHistory: getHistory, memberIDs: memberIDs}
}

// Handlers returns the per-type handlers to register on the realtime router.
func (h *GroupWSHandler) Handlers() map[string]realtime.WSHandler {
	return map[string]realtime.WSHandler{
		realtime.TypeGroupChatSend:    realtime.HandlerFunc(h.handleSend),
		realtime.TypeGroupChatHistory: realtime.HandlerFunc(h.handleHistory),
	}
}

func (h *GroupWSHandler) handleSend(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.GroupChatSendPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendGroupRealtimeError(client, env.RequestID, "invalid group chat send payload")
		return
	}

	result, err := h.send.Execute(context.Background(), commands.SendGroupMessageCommand{
		GroupID:  payload.GroupID,
		SenderID: client.UserID,
		Content:  payload.Content,
	})
	if err != nil {
		sendGroupRealtimeError(client, env.RequestID, err.Error())
		return
	}

	memberIDs, err := h.memberIDs.Resolve(context.Background(), queries.ListGroupMemberIDsQuery{GroupID: payload.GroupID})
	if err != nil {
		sendGroupRealtimeError(client, env.RequestID, err.Error())
		return
	}

	out, _ := json.Marshal(realtime.GroupChatMessagePayload{
		ID:        result.Message.ID,
		GroupID:   result.Message.GroupID,
		SenderID:  result.Message.SenderID,
		Content:   result.Message.Content,
		CreatedAt: result.Message.CreatedAt,
	})
	reply, _ := json.Marshal(realtime.Envelope{Type: realtime.TypeGroupChatMessage, Payload: out})

	h.hub.SendToUsers(memberIDs, reply)
}

func (h *GroupWSHandler) handleHistory(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.GroupChatHistoryPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		sendGroupRealtimeError(client, env.RequestID, "invalid group chat history payload")
		return
	}

	result, err := h.getHistory.Resolve(context.Background(), queries.GetGroupChatQuery{
		GroupID: payload.GroupID,
		UserID:  client.UserID,
		Limit:   payload.Limit,
	})
	if err != nil {
		sendGroupRealtimeError(client, env.RequestID, err.Error())
		return
	}

	out, _ := json.Marshal(result.Messages)
	reply, _ := json.Marshal(realtime.Envelope{
		Type:      realtime.TypeGroupChatHistResult,
		RequestID: env.RequestID,
		Payload:   out,
	})
	client.Send(reply)
}

func sendGroupRealtimeError(client *realtime.Client, requestID, message string) {
	payload, _ := json.Marshal(realtime.ErrorPayload{Message: message})
	reply, _ := json.Marshal(realtime.Envelope{
		Type:      realtime.TypeError,
		RequestID: requestID,
		Payload:   payload,
	})
	client.Send(reply)
}
