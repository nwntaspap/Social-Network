package transport

import (
	"errors"
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/topic"
	"social-network/internal/topic/commands"
)

func (h *Handler) CastVote(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUser(r)
	if !ok {
		h.logger.PrintError(errors.New("user not authenticated"), nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	topicID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		h.logger.PrintError(errors.New("invalid topic ID"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid topic ID")
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req struct {
			ReactionType int `json:"reactionType"`
		}
		if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
			h.logger.PrintError(errors.New("invalid request payload"), nil)
			helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		if err := h.castVote.Execute(r.Context(), commands.CastVoteCommand{
			UserID:       userID,
			TopicID:      topicID,
			ReactionType: req.ReactionType,
		}); err != nil {
			h.logger.PrintError(err, nil)
			switch {
			case errors.Is(err, topic.ErrInvalidVoteValue):
				helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, topic.ErrTopicNotFound):
				helpers.RespondWithError(w, http.StatusNotFound, err.Error())
			default:
				helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Vote cast successfully"})

	case http.MethodDelete:
		if err := h.deleteVote.Execute(r.Context(), commands.DeleteVoteCommand{
			UserID:  userID,
			TopicID: topicID,
		}); err != nil {
			if !errors.Is(err, topic.ErrTopicNotFound) {
				h.logger.PrintError(err, nil)
				helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Vote removed"})

	default:
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
	}
}
