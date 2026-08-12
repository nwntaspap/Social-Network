package transport

import (
	"errors"
	"net/http"

	"social-network/internal/comment/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetCommentByIDWithVotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	commentID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		h.logger.PrintError(errors.New("invalid comment ID"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		h.logger.PrintError(errors.New("user not authenticated"), nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	c, err := h.getCommentWV.Resolve(r.Context(), queries.GetCommentByIDWithVotesQuery{
		CommentID: commentID,
		UserID:    userID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toCommentResponse(c, h.lookupUser(r.Context(), c.UserID))
	helpers.RespondWithJSON(w, http.StatusOK, nil, resp)
}
