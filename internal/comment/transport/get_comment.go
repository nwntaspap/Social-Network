package transport

import (
	"errors"
	"net/http"

	"social-network/internal/comment/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetCommentByID(w http.ResponseWriter, r *http.Request) {
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

	c, err := h.getComment.Resolve(r.Context(), queries.GetCommentByIDQuery{
		CommentID: commentID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toCommentResponse(c, h.lookupUser(r.Context(), c.UserID))
	helpers.RespondWithJSON(w, http.StatusOK, nil, resp)
}
