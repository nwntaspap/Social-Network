package transport

import (
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/topic/commands"
)

func (h *Handler) CastVote(w http.ResponseWriter, r *http.Request) {
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

	switch r.Method {
	case http.MethodPost:
		if err := h.castVote.Execute(r.Context(), commands.CastVoteCommand{
			UserID:       userID,
			TopicID:      topicID,
			ReactionType: 1,
		}); err != nil {
			helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Post liked"})

	case http.MethodDelete:
		if err := h.castVote.Execute(r.Context(), commands.CastVoteCommand{
			UserID:       userID,
			TopicID:      topicID,
			ReactionType: -1,
		}); err != nil {
			helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Post unliked"})

	default:
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
	}
}
