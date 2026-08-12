package transport

import (
	"context"
	"net/http"

	"social-network/internal/chat"
	"social-network/internal/chat/commands"
	"social-network/internal/chat/queries"
	"social-network/internal/platform/logger"
)

type UserExtractor func(r *http.Request) (userID string, ok bool)

type UserResult struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Nickname    string `json:"nickname,omitempty"`
	AboutMe     string `json:"aboutMe,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	DateOfBirth string `json:"dateOfBirth"`
	IsPublic    bool   `json:"isPublic"`
	CreatedAt   string `json:"createdAt"`
}

// UserLookup is a local interface for nested user lookups (avoids importing domain/user).
type UserLookup interface {
	GetUserByID(ctx context.Context, id string) (*UserResult, error)
}

type ChatHistoryResolver interface {
	Resolve(ctx context.Context, q queries.GetChatHistoryQuery) ([]*chat.Message, error)
}

type ChatUsersResolver interface {
	Resolve(ctx context.Context, q queries.GetChatUsersRequest) ([]queries.Conversation, error)
}

type StartChatExecutor interface {
	Execute(ctx context.Context, cmd commands.OpenPrivateChatCommand) (commands.OpenPrivateChatResult, error)
}

type Handler struct {
	logger      logger.Logger
	getHistory  ChatHistoryResolver
	getUsers    ChatUsersResolver
	startChat   StartChatExecutor
	extractUser UserExtractor
	userLookup  UserLookup
}

func NewHandler(
	logger logger.Logger,
	extract UserExtractor,
	userLookup UserLookup,
	getHistory ChatHistoryResolver,
	getUsers ChatUsersResolver,
) *Handler {
	return &Handler{
		logger:      logger,
		getHistory:  getHistory,
		getUsers:    getUsers,
		extractUser: extract,
		userLookup:  userLookup,
	}
}

func NewHandlerWithStart(
	logger logger.Logger,
	extract UserExtractor,
	userLookup UserLookup,
	getHistory ChatHistoryResolver,
	getUsers ChatUsersResolver,
	startChat StartChatExecutor,
) *Handler {
	h := NewHandler(logger, extract, userLookup, getHistory, getUsers)
	h.startChat = startChat
	return h
}

func (h *Handler) lookupUser(ctx context.Context, userID string) *UserResult {
	if h.userLookup == nil || userID == "" {
		return nil
	}
	u, err := h.userLookup.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}
	return u
}
