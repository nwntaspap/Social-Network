package transport

import (
	"net/http"
	"strconv"

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

	helpers.RespondWithJSON(w, http.StatusOK, nil, result)
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

	_ = userID

	messages, err := h.getHistory.Resolve(r.Context(), queries.GetChatHistoryQuery{
		ChatID:          chatID,
		BeforeMessageID: beforeID,
		Limit:           limit,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to get chat history")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, messages)
}
