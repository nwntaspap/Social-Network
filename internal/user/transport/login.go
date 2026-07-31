package transport

import (
	"errors"
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user/commands"
)

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	var req loginRequest
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.login.Execute(r.Context(), commands.LoginCommand{
		Identifier: req.Identifier,
		Password:   req.Password,
	})
	if err != nil {
		if errors.Is(err, commands.ErrInvalidCredentials) {
			helpers.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, "login failed")
		return
	}

	h.sessionCookies.SetAccessCookie(w, result.Token, result.ExpiresAt)

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]any{
		"token": result.Token,
		"user":  userResponse(result.User),
	})
}
