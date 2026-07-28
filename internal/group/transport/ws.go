package transport

import (
	"encoding/json"
	"net/http"

	"social-network/internal/pkg/helpers"
)

type GroupChatEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type GroupChatPayload struct {
	GroupID string `json:"groupId"`
	Content string `json:"content"`
}

const TypeGroupChatMessage = "group_chat_message"

func (h *Handler) HandleGroupChatWS(w http.ResponseWriter, r *http.Request) {
	groupID, err := helpers.GetQueryString(r, "groupId")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "groupId is required")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	_ = groupID
	_ = userID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","message":"group chat websocket placeholder"}`))
}
