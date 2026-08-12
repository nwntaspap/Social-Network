package transport

import (
	"net/http"

	"social-network/internal/comment/commands"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) DeleteCommentVote(w http.ResponseWriter, r *http.Request) {
	userID, commentID, ok := h.requireCommentDeleteContext(w, r)
	if !ok {
		return
	}

	if err := h.deleteCommentVote.Execute(r.Context(), commands.DeleteCommentVoteCommand{
		UserID:    userID,
		CommentID: commentID,
	}); err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Vote deleted successfully"})
}
