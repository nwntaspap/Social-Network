package transport

import (
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user/queries"
)

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	userID, ok := h.auth.Extract(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: user not found")
		return
	}

	result, err := h.getProfile.Resolve(r.Context(), queries.GetProfileQuery{
		TargetID:    userID,
		RequesterID: userID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, userResponse(&result.User))
}
