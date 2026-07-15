package transport

import (
	"errors"
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user"
	"social-network/internal/user/queries"
)

func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	userID, err := helpers.GetQueryString(r, "user_id")
	if err != nil || userID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}

	result, err := h.getActivity.Resolve(r.Context(), queries.GetActivityQuery{UserID: userID})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			helpers.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to get activity")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]any{
		"postCount":      result.PostCount,
		"commentCount":   result.CommentCount,
		"voteCount":      result.VoteCount,
		"followerCount":  result.FollowerCount,
		"followingCount": result.FollowingCount,
	})
}
