package userlogin

import (
	"context"
	"net/http"

	"social-network/internal/infra/logger"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/validator"

	userQueries "social-network/internal/app/user/queries"
)

type LoginUserUsernameRequestModel struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h Handler) UserLoginUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.Logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.Config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	var userToLogin LoginUserUsernameRequestModel

	userAny, err := helpers.ParseBodyRequest(r, &userToLogin)
	if err != nil {
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			"invalid request: "+err.Error(),
		)

		h.Logger.PrintError(err, nil)

		return
	}
	defer r.Body.Close()

	v := validator.New()

	validator.ValidateUserLoginUsername(v, userAny)

	if !v.Valid() {
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			v.ToStringErrors(),
		)

		h.Logger.PrintError(logger.ErrValidationFailed, v.Errors)

		return
	}

	user, err := h.UserServices.Queries.UserLoginUsername.Handle(ctx, userQueries.UserLoginUsernameRequest{
		Username: userToLogin.Username,
		Password: userToLogin.Password,
	})
	if err != nil {
		helpers.RespondWithError(
			w,
			http.StatusInternalServerError,
			"error logging in user",
		)

		h.Logger.PrintError(err, nil)
		return
	}

	newSession, err := h.SessionManager.CreateSession(ctx, user.ID)
	if err != nil {
		helpers.RespondWithError(
			w,
			http.StatusInternalServerError,
			"error creating session",
		)

		h.Logger.PrintError(err, nil)
		return
	}

	loginResponse := LoginResponse{
		UserID:   user.ID,
		Username: user.Nickname,
		// AccessToken:  newSession.AccessToken,
		// RefreshToken: newSession.RefreshToken,
		// not sure if we need them anymore in the response
	}

	h.CookieManager.SetCookies(w, newSession)

	helpers.RespondWithJSON(
		w,
		http.StatusOK,
		nil,
		loginResponse,
	)

	h.Logger.PrintInfo(
		"User login successfully",
		map[string]string{
			"userId": user.ID,
			"name":   user.Nickname,
		},
	)
}
