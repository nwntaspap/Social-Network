package transport

import (
	"errors"
	"net/http"

	"social-network/internal/core/middleware"
	"social-network/internal/pkg/helpers"
	"social-network/internal/user/commands"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	token := middleware.GetSessionTokenFromContext(r)
	if token == "" {
		h.logger.PrintError(errors.New("no active session"), nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "no active session")
		return
	}

	if err := h.logout.Execute(r.Context(), commands.LogoutCommand{Token: token}); err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	h.sessionCookies.DeleteAccessCookie(w)
	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"message": "logged out"})
}
