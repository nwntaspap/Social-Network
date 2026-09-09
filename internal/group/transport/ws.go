package transport

import (
	"context"
	"encoding/json"
	"slices"

	"social-network/internal/core/realtime"
	"social-network/internal/group/commands"
	"social-network/internal/group/queries"
	"social-network/internal/platform/logger"
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
	markRead   MarkGroupReadExecutor
	logger     logger.Logger
}

// MarkGroupReadExecutor marks a group's chat as read by the user.
type MarkGroupReadExecutor interface {
	Execute(ctx context.Context, cmd commands.MarkGroupReadCommand) error
}

func NewGroupWSHandler(
	hub *realtime.Hub,
	send SendGroupMessageExecutor,
	getHistory GetGroupChatResolver,
	memberIDs GroupMemberIDsResolver,
	markRead MarkGroupReadExecutor,
	logger logger.Logger,
) *GroupWSHandler {
	return &GroupWSHandler{hub: hub, send: send, getHistory: getHistory, memberIDs: memberIDs, markRead: markRead, logger: logger}
}

// Handlers returns the per-type handlers to register on the realtime router.
func (h *GroupWSHandler) Handlers() map[string]realtime.WSHandler {
	return map[string]realtime.WSHandler{
		realtime.TypeGroupChatSend:     realtime.HandlerFunc(h.handleSend),
		realtime.TypeGroupChatHistory:  realtime.HandlerFunc(h.handleHistory),
		realtime.TypeGroupChatMarkRead: realtime.HandlerFunc(h.handleMarkRead),
		realtime.TypeGroupChatTyping:   realtime.HandlerFunc(h.handleTyping),
	}
}

func (h *GroupWSHandler) handleSend(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.GroupChatSendPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		h.logger.PrintError(err, nil)
		sendGroupRealtimeError(client, env.RequestID, "invalid group chat send payload")
		return
	}

	result, err := h.send.Execute(context.Background(), commands.SendGroupMessageCommand{
		GroupID:  payload.GroupID,
		SenderID: client.UserID,
		Content:  payload.Content,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		sendGroupRealtimeError(client, env.RequestID, err.Error())
		return
	}

	memberIDs, err := h.memberIDs.Resolve(context.Background(), queries.ListGroupMemberIDsQuery{GroupID: payload.GroupID})
	if err != nil {
		h.logger.PrintError(err, nil)
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
		h.logger.PrintError(err, nil)
		sendGroupRealtimeError(client, env.RequestID, "invalid group chat history payload")
		return
	}

	result, err := h.getHistory.Resolve(context.Background(), queries.GetGroupChatQuery{
		GroupID: payload.GroupID,
		UserID:  client.UserID,
		Limit:   payload.Limit,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
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

func (h *GroupWSHandler) handleMarkRead(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.GroupChatMarkReadPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		h.logger.PrintError(err, nil)
		sendGroupRealtimeError(client, env.RequestID, "invalid group_chat.mark_read payload")
		return
	}
	if err := h.markRead.Execute(context.Background(), commands.MarkGroupReadCommand{
		GroupID: payload.GroupID,
		UserID:  client.UserID,
	}); err != nil {
		h.logger.PrintError(err, nil)
		sendGroupRealtimeError(client, env.RequestID, err.Error())
	}
}

func (h *GroupWSHandler) handleTyping(client *realtime.Client, env realtime.Envelope) {
	var payload realtime.GroupChatTypingPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		h.logger.PrintError(err, nil)
		sendGroupRealtimeError(client, env.RequestID, "invalid group_chat.typing payload")
		return
	}

	memberIDs, err := h.memberIDs.Resolve(context.Background(), queries.ListGroupMemberIDsQuery{GroupID: payload.GroupID})
	if err != nil {
		h.logger.PrintError(err, nil)
		sendGroupRealtimeError(client, env.RequestID, err.Error())
		return
	}

	if !slices.Contains(memberIDs, client.UserID) {
		sendGroupRealtimeError(client, env.RequestID, "you are not a member of this group")
		return
	}

	out, _ := json.Marshal(realtime.GroupChatIsTyping{GroupID: payload.GroupID, UserID: client.UserID})
	reply, _ := json.Marshal(realtime.Envelope{Type: realtime.TypeGroupIsTyping, Payload: out})

	recipients := make([]string, 0, len(memberIDs))
	for _, id := range memberIDs {
		if id != client.UserID {
			recipients = append(recipients, id)
		}
	}
	h.hub.SendToUsers(recipients, reply)
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
