package transport

import (
	"net/http"

	"social-network/internal/follow/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, err := helpers.GetQueryString(r, "userId")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "userId query param required")
		return
	}

	followers, err := h.getFollowers.Resolve(r.Context(), queries.GetFollowersQuery{UserID: userID})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, followers)
}
