package transport

import (
	"net/http"

	"social-network/internal/comment/commands"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	commentID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	if err := h.deleteComment.Execute(r.Context(), commands.DeleteCommentCommand{
		UserID:    userID,
		CommentID: commentID,
	}); err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Comment deleted successfully"})
}
