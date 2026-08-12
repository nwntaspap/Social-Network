package transport

import (
	"errors"
	"net/http"

	"social-network/internal/comment/commands"
	"social-network/internal/comment/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) CastCommentVote(w http.ResponseWriter, r *http.Request) {
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

	commentID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		h.logger.PrintError(errors.New("invalid comment ID"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid comment ID")
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

	if err := h.castCommentVote.Execute(r.Context(), commands.CastCommentVoteCommand{
		UserID:       userID,
		CommentID:    commentID,
		ReactionType: req.ReactionType,
	}); err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Vote cast successfully"})
}

func (h *Handler) GetVoteCounts(w http.ResponseWriter, r *http.Request) {
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

	counts, err := h.getCommentVotes.Resolve(r.Context(), queries.GetVoteCountsQuery{
		CommentID: commentID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := VoteCountsResponse{
		Upvotes:   counts.Upvotes,
		Downvotes: counts.Downvotes,
		Score:     counts.Score,
	}
	helpers.RespondWithJSON(w, http.StatusOK, nil, resp)
}
