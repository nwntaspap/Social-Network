package transport

import (
	"errors"
	"net/http"

	"social-network/internal/follow/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, err := helpers.GetQueryString(r, "userId")
	if err != nil {
		h.logger.PrintError(errors.New("userId query param required"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "userId query param required")
		return
	}

	followers, err := h.getFollowers.Resolve(r.Context(), queries.GetFollowersQuery{UserID: userID})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	users := make([]*UserResult, len(followers))
	for i, f := range followers {
		users[i] = h.lookupUser(r.Context(), f.FollowerID)
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, users)
}
