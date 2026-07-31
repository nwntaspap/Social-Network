package transport

import (
	"net/http"

	"social-network/internal/comment/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetCommentsByTopicWithVotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	topicID, err := helpers.GetQueryInt(r, "topicId")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid topic ID")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	comments, err := h.getByTopicWV.Resolve(r.Context(), queries.GetCommentsByTopicWithVotesQuery{
		TopicID: topicID,
		UserID:  userID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = toCommentResponse(&c, h.lookupUser(r.Context(), c.UserID))
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, responses)
}
