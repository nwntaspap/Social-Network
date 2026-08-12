package transport

import (
	"errors"
	"net/http"

	"social-network/internal/follow/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	requests, err := h.getPendingReqs.Resolve(r.Context(), queries.GetPendingRequestsQuery{
		UserID: userID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]FollowRequestResponse, len(requests))
	for i, req := range requests {
		responses[i] = toFollowRequestResponse(
			req,
			h.lookupUser(r.Context(), req.FollowerID),
			h.lookupUser(r.Context(), req.FolloweeID),
		)
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, responses)
}
