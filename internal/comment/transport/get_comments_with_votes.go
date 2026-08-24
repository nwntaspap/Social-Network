package transport

import (
	"errors"
	"net/http"

	"social-network/internal/comment/queries"
	"social-network/internal/pkg/helpers"
	"social-network/internal/topic"
)

func (h *Handler) GetCommentsByTopicWithVotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	topicID, err := helpers.GetQueryInt(r, "topicId")
	if err != nil {
		h.logger.PrintError(errors.New("invalid topic ID"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid topic ID")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		h.logger.PrintError(errors.New("user not authenticated"), nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	comments, err := h.getByTopicWV.Resolve(r.Context(), queries.GetCommentsByTopicWithVotesQuery{
		TopicID:     topicID,
		RequesterID: userID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		if errors.Is(err, topic.ErrTopicNotFound) {
			helpers.RespondWithError(w, http.StatusNotFound, "Topic not found")
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = toCommentResponse(&c, h.lookupUser(r.Context(), c.UserID))
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, responses)
}
