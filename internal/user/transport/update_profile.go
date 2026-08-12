package transport

import (
	"errors"
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user"
	"social-network/internal/user/commands"
)

type updateProfileRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Nickname  string `json:"nickname"`
	AboutMe   string `json:"aboutMe"`
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	userID, ok := h.auth.Extract(r)
	if !ok {
		h.logger.PrintError(errors.New("unauthorized"), nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateProfileRequest
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		h.logger.PrintError(errors.New("invalid request body"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.updateProfile.Execute(r.Context(), commands.UpdateProfileCommand{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Nickname:  req.Nickname,
		AboutMe:   req.AboutMe,
	}); err != nil {
		h.logger.PrintError(err, nil)
		if errors.Is(err, user.ErrUserNotFound) {
			helpers.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "profile updated"})
}
