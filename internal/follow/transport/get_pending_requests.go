package transport

import (
	"net/http"

	"social-network/internal/follow/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	requests, err := h.getPendingReqs.Resolve(r.Context(), queries.GetPendingRequestsQuery{
		UserID: userID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, requests)
}
