package transport

import (
	"errors"
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user"
	"social-network/internal/user/queries"
)

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	targetID, err := helpers.GetQueryString(r, "user_id")
	if err != nil || targetID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}

	requesterID := ""
	if id, ok := h.auth.Extract(r); ok {
		requesterID = id
	}

	result, err := h.getProfile.Resolve(r.Context(), queries.GetProfileQuery{
		TargetID:    targetID,
		RequesterID: requesterID,
	})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			helpers.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to get profile")
		return
	}

	resp := userResponse(&result.User)
	resp["followersCount"] = result.FollowerCount
	resp["followingCount"] = result.FollowingCount
	resp["isFollowing"] = result.IsFollowing

	helpers.RespondWithJSON(w, http.StatusOK, nil, resp)
}
