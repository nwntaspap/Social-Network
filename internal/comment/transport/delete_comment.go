package transport

import (
	"errors"
	"net/http"

	"social-network/internal/comment/commands"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) requireCommentDeleteContext(w http.ResponseWriter, r *http.Request) (userID string, commentID int, ok bool) {
	if r.Method != http.MethodDelete {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return "", 0, false
	}

	userID, ok = h.extractUser(r)
	if !ok {
		h.logger.PrintError(errors.New("user not authenticated"), nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return "", 0, false
	}

	commentID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		h.logger.PrintError(errors.New("invalid comment ID"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid comment ID")
		return "", 0, false
	}

	return userID, commentID, true
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, commentID, ok := h.requireCommentDeleteContext(w, r)
	if !ok {
		return
	}

	if err := h.deleteComment.Execute(r.Context(), commands.DeleteCommentCommand{
		UserID:    userID,
		CommentID: commentID,
	}); err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Comment deleted successfully"})
}
