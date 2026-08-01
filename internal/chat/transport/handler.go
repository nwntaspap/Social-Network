package transport

import (
	"context"
	"net/http"

	"social-network/internal/chat"
	"social-network/internal/chat/queries"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type ChatHistoryResolver interface {
	Resolve(ctx context.Context, q queries.GetChatHistoryQuery) ([]*chat.Message, error)
}

type ChatUsersResolver interface {
	Resolve(ctx context.Context, q queries.GetChatUsersRequest) ([]queries.Conversation, error)
}

type Handler struct {
	getHistory  ChatHistoryResolver
	getUsers    ChatUsersResolver
	extractUser UserExtractor
}

func NewHandler(
	extract UserExtractor,
	getHistory ChatHistoryResolver,
	getUsers ChatUsersResolver,
) *Handler {
	return &Handler{
		getHistory:  getHistory,
		getUsers:    getUsers,
		extractUser: extract,
	}
}
