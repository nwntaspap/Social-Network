package transport

import (
	"net/http"

	"social-network/internal/comment/commands"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		CommentID int    `json:"commentId"`
		Content   string `json:"content"`
	}
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := h.updateComment.Execute(r.Context(), commands.UpdateCommentCommand{
		UserID:    userID,
		CommentID: req.CommentID,
		Content:   req.Content,
	}); err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Comment updated successfully"})
}
