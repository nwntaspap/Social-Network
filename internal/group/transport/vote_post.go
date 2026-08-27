package transport

import (
	"errors"
	"net/http"

	"social-network/internal/group/commands"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) VoteGroupPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		h.logger.PrintError(errors.New("user not authenticated"), nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	postID := r.PathValue("postId")
	if postID == "" {
		h.logger.PrintError(errors.New("postId is required"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "postId is required")
		return
	}

	var req struct {
		ReactionType int `json:"reactionType"`
	}
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		h.logger.PrintError(errors.New("invalid request payload"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := h.castGroupPostVote.Execute(r.Context(), commands.CastGroupPostVoteCommand{
		UserID:       userID,
		PostID:       postID,
		ReactionType: req.ReactionType,
	}); err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Vote updated"})
}
