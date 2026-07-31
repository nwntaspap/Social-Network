package transport

import (
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/topic/queries"
)

func (h *Handler) GetTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	topicID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid topic ID")
		return
	}

	var userID *string
	if uid, ok := h.extractUser(r); ok {
		userID = &uid
	}

	top, err := h.getTopic.Resolve(r.Context(), queries.GetTopicQuery{
		TopicID: topicID,
		UserID:  userID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toTopicResponse(top, h.lookupUser(r.Context(), top.UserID))
	helpers.RespondWithJSON(w, http.StatusOK, nil, resp)
}
