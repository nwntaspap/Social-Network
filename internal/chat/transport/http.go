package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"social-network/internal/chat"
	"social-network/internal/chat/commands"
	"social-network/internal/chat/queries"
	"social-network/internal/pkg/helpers"
	"social-network/internal/platform/logger"
)

func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.PrintError(logger.ErrMethodNotAllowed, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		h.logger.PrintError(logger.ErrUserNotFoundInContext, nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	result, err := h.getUsers.Resolve(r.Context(), queries.GetChatUsersRequest{MeID: userID})
	if err != nil {
		h.logger.PrintError(errors.New("failed to get conversations"), nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to get conversations")
		return
	}

	responses := make([]ConversationResponse, len(result))
	for i, c := range result {
		responses[i] = toConversationResponse(c)
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, responses)
}

func (h *Handler) GetChatHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.PrintError(logger.ErrMethodNotAllowed, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		h.logger.PrintError(logger.ErrUserNotFoundInContext, nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	chatID := r.URL.Query().Get("chatId")
	if chatID == "" {
		h.logger.PrintError(errors.New("ChatId is required"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "chatId is required")
		return
	}

	beforeID, _ := strconv.Atoi(r.URL.Query().Get("before"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	messages, err := h.getHistory.Resolve(r.Context(), queries.GetChatHistoryQuery{
		ChatID:          chatID,
		RequesterID:     userID,
		BeforeMessageID: beforeID,
		Limit:           limit,
	})
	if err != nil {
		switch {
		case errors.Is(err, chat.ErrNotParticipant):
			h.logger.PrintError(err, nil)
			helpers.RespondWithError(w, http.StatusForbidden, "You are not a participant of this chat")
		case errors.Is(err, chat.ErrNotConnected):
			h.logger.PrintError(err, nil)
			helpers.RespondWithError(w, http.StatusForbidden, "You are not connected to this user")
		case errors.Is(err, chat.ErrChatNotFound):
			h.logger.PrintError(err, nil)
			helpers.RespondWithError(w, http.StatusNotFound, "Chat not found")
		default:
			h.logger.PrintError(err, nil)
			helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to get chat history")
		}
		return
	}

	responses := make([]ChatMessageResponse, len(messages))
	for i, m := range messages {
		responses[i] = toMessageResponse(m, h.lookupUser(r.Context(), m.SenderID))
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, responses)
}

func (h *Handler) StartChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.PrintError(logger.ErrMethodNotAllowed, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		h.logger.PrintError(logger.ErrUserNotFoundInContext, nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if h.startChat == nil {
		h.logger.PrintError(errors.New("start chat is not available"), nil)
		helpers.RespondWithError(w, http.StatusNotImplemented, "Start chat is not available")
		return
	}

	var body struct {
		UserID string `json:"userId"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		h.logger.PrintError(logger.ErrInvalidRequestBody, nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	body.UserID = strings.TrimSpace(body.UserID)
	if body.UserID == "" {
		h.logger.PrintError(errors.New("UserID is required"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "userId is required")
		return
	}
	if body.UserID == userID {
		h.logger.PrintError(errors.New("you cannot start a chat with yourself"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "You cannot start a chat with yourself")
		return
	}

	result, err := h.startChat.Execute(r.Context(), commands.OpenPrivateChatCommand{
		SenderID:   userID,
		ReceiverID: body.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, commands.ErrNotConnected):
			h.logger.PrintError(err, nil)
			helpers.RespondWithError(w, http.StatusForbidden, "You are not connected to this user")
		case errors.Is(err, commands.ErrCannotMessage):
			h.logger.PrintError(err, nil)
			helpers.RespondWithError(w, http.StatusForbidden, "This user cannot receive your messages")
		default:
			h.logger.PrintError(err, nil)
			helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to start chat")
		}
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, result.Chat)
}
