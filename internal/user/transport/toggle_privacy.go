package transport

import (
	"errors"
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user"
	"social-network/internal/user/commands"
)

type togglePrivacyRequest struct {
	IsPrivate bool `json:"isPrivate"`
}

func (h *Handler) TogglePrivacy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	userID, ok := h.auth.Extract(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req togglePrivacyRequest
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.togglePrivacy.Execute(r.Context(), commands.TogglePrivacyCommand{
		UserID:    userID,
		IsPrivate: req.IsPrivate,
	}); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			helpers.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to toggle privacy")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "privacy updated"})
}
