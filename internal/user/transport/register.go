package transport

import (
	"errors"
	"net/http"
	"time"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user/commands"
)

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Nickname    string `json:"nickname"`
	DateOfBirth string `json:"dateOfBirth"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	var req registerRequest
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dob, err := time.Parse(time.RFC3339, req.DateOfBirth)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid dateOfBirth format, use RFC3339")
		return
	}

	u, err := h.register.Execute(r.Context(), commands.RegisterCommand{
		Email:       req.Email,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Nickname:    req.Nickname,
		DateOfBirth: dob,
	})
	if err != nil {
		switch {
		case errors.Is(err, commands.ErrEmailTaken):
			helpers.RespondWithError(w, http.StatusConflict, err.Error())
		case errors.Is(err, commands.ErrUnderage),
			errors.Is(err, commands.ErrWeakPassword),
			errors.Is(err, commands.ErrNicknameEmpty):
			helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			helpers.RespondWithError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	helpers.RespondWithJSON(w, http.StatusCreated, nil, map[string]any{
		"id":    u.ID,
		"email": u.Email,
	})
}
