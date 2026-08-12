package transport

import (
	"errors"
	"net/http"

	"social-network/internal/comment/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetCommentsByTopic(w http.ResponseWriter, r *http.Request) {
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

	comments, err := h.getByTopic.Resolve(r.Context(), queries.GetCommentsByTopicQuery{
		TopicID: topicID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = toCommentResponse(&c, h.lookupUser(r.Context(), c.UserID))
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, responses)
}
