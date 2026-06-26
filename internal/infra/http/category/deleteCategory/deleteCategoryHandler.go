package deletecategory

import (
	"context"
	"net/http"
	"strconv"

	"social-network/internal/app"
	"social-network/internal/config"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/validator"

	categorycommands "social-network/internal/app/categories/commands"
)

type ResponseModel struct {
	Message    string `json:"message"`
	CategoryID int    `json:"categoryId"`
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

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.Logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.Config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	user := middleware.GetUserFromContext(r)
	if user == nil {
		h.Logger.PrintError(logger.ErrUserNotFoundInContext, nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	categoryID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
	}

	val := validator.New()
	validator.ValidateDeleteCategory(val, &struct {
		CategoryID int
	}{
		CategoryID: categoryID,
	})

	if !val.Valid() {
		h.Logger.PrintError(logger.ErrValidationFailed, val.Errors)
		helpers.RespondWithError(w,
			http.StatusBadRequest,
			val.ToStringErrors())
		return
	}

	err = h.UserServices.Commands.DeleteCategory.Handle(ctx, categorycommands.DeleteCategoryRequest{
		CategoryID: categoryID,
		UserID:     user.ID,
	})
	if err != nil {
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Error deleting category")
		return
	}

	helpers.RespondWithJSON(w,
		http.StatusOK,
		nil,
		ResponseModel{
			CategoryID: categoryID,
			Message:    "Category deleted successfully",
		})

	h.Logger.PrintInfo(
		"Category deleted successfully",
		map[string]string{
			"cat_id":  strconv.Itoa(categoryID),
			"user_id": user.ID,
		},
	)
}
