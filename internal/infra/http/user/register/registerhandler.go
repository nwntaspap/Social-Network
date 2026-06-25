package userregister

import (
	"context"
	"net/http"
	"strings"

	"social-network/internal/app"
	"social-network/internal/config"
	"social-network/internal/domain/session"
	"social-network/internal/infra/logger"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/validator"

	usercommands "social-network/internal/app/user/commands"
)

type RegisterUserReguestModel struct {
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Gender    string `json:"gender"`
	Age       int    `json:"age"`
}

type RegisterUserResponse struct {
	UserID  string `json:"userId"`
	Message string `json:"message"`
}

type Handler struct {
	UserServices   app.Services
	SessionManager session.Manager
	Config         *config.ServerConfig
	Logger         logger.Logger
}

func NewHandler(config *config.ServerConfig, app app.Services, sm session.Manager, logger logger.Logger) *Handler {
	return &Handler{
		UserServices:   app,
		SessionManager: sm,
		Config:         config,
		Logger:         logger,
	}
}

func (h Handler) UserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.Logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.Config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	var userToRegister RegisterUserReguestModel

	userAny, err := helpers.ParseBodyRequest(r, &userToRegister)
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

	validator.ValidateUserRegistration(v, userAny)

	if !v.Valid() {
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			v.ToStringErrors(),
		)

		h.Logger.PrintError(logger.ErrValidationFailed, v.Errors)
		return
	}

	user, err := h.UserServices.Commands.UserRegister.Handle(ctx, usercommands.UserRegisterRequest{
		Nickname:  userToRegister.Nickname,
		Password:  userToRegister.Password,
		FirstName: userToRegister.FirstName,
		LastName:  userToRegister.LastName,
		Age:       userToRegister.Age,
		Gender:    userToRegister.Gender,
		Email:     strings.ToLower(userToRegister.Email),
	})
	if err != nil {
		helpers.RespondWithError(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)

		h.Logger.PrintError(err, nil)

		return
	}

	userResponse := RegisterUserResponse{
		UserID:  user.ID,
		Message: "user registered successfully",
	}

	helpers.RespondWithJSON(
		w,
		http.StatusCreated,
		nil,
		userResponse,
	)

	h.Logger.PrintInfo(
		"User registered successfully",
		map[string]string{
			"userId": user.ID,
			"email":  user.Email,
			"name":   user.Nickname,
		},
	)
}
