package transport

import (
	"net/http"

	"social-network/internal/core/middleware"
	"social-network/internal/pkg/helpers"
	"social-network/internal/user/commands"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	token := middleware.GetSessionTokenFromContext(r)
	if token == "" {
		helpers.RespondWithError(w, http.StatusUnauthorized, "no active session")
		return
	}

	if err := h.logout.Execute(r.Context(), commands.LogoutCommand{Token: token}); err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "logged out"})
}
