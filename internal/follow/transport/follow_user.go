package transport

import (
	"errors"
	"net/http"

	"social-network/internal/follow/commands"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) FollowUser(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		TargetID string `json:"targetId"`
	}
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		h.logger.PrintError(errors.New("invalid request payload"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	result, err := h.followUser.Execute(r.Context(), commands.FollowUserCommand{
		FollowerID: userID,
		TargetID:   req.TargetID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{
		"status": string(result),
	})
}
