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
)

func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	result, err := h.getUsers.Resolve(r.Context(), queries.GetChatUsersRequest{MeID: userID})
	if err != nil {
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
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	chatID := r.URL.Query().Get("chatId")
	if chatID == "" {
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
			helpers.RespondWithError(w, http.StatusForbidden, "You are not a participant of this chat")
		case errors.Is(err, chat.ErrNotConnected):
			helpers.RespondWithError(w, http.StatusForbidden, "You are not connected to this user")
		case errors.Is(err, chat.ErrChatNotFound):
			helpers.RespondWithError(w, http.StatusNotFound, "Chat not found")
		default:
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
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if h.startChat == nil {
		helpers.RespondWithError(w, http.StatusNotImplemented, "Start chat is not available")
		return
	}

	var body struct {
		UserID string `json:"userId"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	body.UserID = strings.TrimSpace(body.UserID)
	if body.UserID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "userId is required")
		return
	}
	if body.UserID == userID {
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
			helpers.RespondWithError(w, http.StatusForbidden, "You are not connected to this user")
		case errors.Is(err, commands.ErrCannotMessage):
			helpers.RespondWithError(w, http.StatusForbidden, "This user cannot receive your messages")
		default:
			helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to start chat")
		}
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, result.Chat)
}
