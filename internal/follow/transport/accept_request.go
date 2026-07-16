package transport

import (
	"net/http"

	"social-network/internal/follow/commands"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		FollowerID string `json:"followerId"`
	}
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := h.acceptRequest.Execute(r.Context(), commands.AcceptRequestCommand{
		FollowerID: req.FollowerID,
		FolloweeID: userID,
	}); err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "Request accepted"})
}
