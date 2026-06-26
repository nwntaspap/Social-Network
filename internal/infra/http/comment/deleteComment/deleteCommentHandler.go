package deletecomment

import (
	"context"
	"net/http"

	"social-network/internal/app"
	"social-network/internal/config"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/validator"

	commentCommands "social-network/internal/app/comments/commands"
)

type ResponseModel struct {
	Message string `json:"message"`
}

type Handler struct {
	UserServices app.Services
	Config       *config.ServerConfig
	Logger       logger.Logger
}

func NewHandler(userServices app.Services, config *config.ServerConfig, logger logger.Logger) *Handler {
	return &Handler{
		UserServices: userServices,
		Config:       config,
		Logger:       logger,
	}
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.Logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		h.Logger.PrintError(logger.ErrUserNotFoundInContext, nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.Config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	commentID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	val := validator.New()

	commentIDVal := &struct {
		CommentID int
	}{
		CommentID: commentID,
	}
	validator.ValidateDeleteComment(val, commentIDVal)

	if !val.Valid() {
		h.Logger.PrintError(logger.ErrValidationFailed, val.Errors)
		helpers.RespondWithError(w, http.StatusBadRequest, val.ToStringErrors())
		return
	}

	err = h.UserServices.Commands.DeleteComment.Handle(ctx, commentCommands.DeleteCommentRequest{
		CommentID: commentID,
		User:      user,
	})
	if err != nil {
		helpers.RespondWithError(
			w,
			http.StatusInternalServerError,
			"Failed to delete comment",
		)

		h.Logger.PrintError(err, nil)
		return
	}

	commentResponse := ResponseModel{
		Message: "Comment deleted successfully",
	}

	helpers.RespondWithJSON(
		w,
		http.StatusOK,
		nil,
		commentResponse,
	)

	h.Logger.PrintInfo(
		"Comment deleted successfully",
		map[string]string{
			"user_id": user.ID,
		},
	)
}
