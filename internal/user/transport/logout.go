package transport

import (
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user/commands"
)

type logoutRequest struct {
	Token string `json:"token"`
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	var req logoutRequest
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.logout.Execute(r.Context(), commands.LogoutCommand{Token: req.Token}); err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "logged out"})
}
