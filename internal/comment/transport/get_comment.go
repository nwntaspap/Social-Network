package transport

import (
	"net/http"

	"social-network/internal/comment/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetCommentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	commentID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	var userID *string
	if uid, ok := h.extractUser(r); ok {
		userID = &uid
	}

	c, err := h.getComment.Resolve(r.Context(), queries.GetCommentByIDQuery{
		CommentID: commentID,
		UserID:    userID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toCommentResponse(c)
	helpers.RespondWithJSON(w, http.StatusOK, nil, resp)
}
