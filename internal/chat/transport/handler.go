package transport

import (
	"net/http"

	"social-network/internal/chat/commands"
	"social-network/internal/chat/queries"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type Handler struct {
	sendMsg     *commands.SendPrivateMessageHandler
	markRead    *commands.MarkAsReadHandler
	getHistory  *queries.GetChatHistoryResolver
	getUsers    *queries.GetChatUsersResolver
	extractUser UserExtractor
}

func NewHandler(
	extract UserExtractor,
	sendMsg *commands.SendPrivateMessageHandler,
	markRead *commands.MarkAsReadHandler,
	getHistory *queries.GetChatHistoryResolver,
	getUsers *queries.GetChatUsersResolver,
) *Handler {
	return &Handler{
		sendMsg:     sendMsg,
		markRead:    markRead,
		getHistory:  getHistory,
		getUsers:    getUsers,
		extractUser: extract,
	}
}

func (h *Handler) SendMsgHandler() *commands.SendPrivateMessageHandler {
	return h.sendMsg
}

func (h *Handler) MarkReadHandler() *commands.MarkAsReadHandler {
	return h.markRead
}

func (h *Handler) GetHistoryResolver() *queries.GetChatHistoryResolver {
	return h.getHistory
}

func (h *Handler) GetUsersResolver() *queries.GetChatUsersResolver {
	return h.getUsers
}
