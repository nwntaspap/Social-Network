package transport

import (
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/topic/commands"
)

func (h *Handler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	topicID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid topic ID")
		return
	}

	if err := h.deleteTopic.Execute(r.Context(), commands.DeleteTopicCommand{
		TopicID: topicID,
		UserID:  userID,
	}); err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Post deleted successfully"})
}
