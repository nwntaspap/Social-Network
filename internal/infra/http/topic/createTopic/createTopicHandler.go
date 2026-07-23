package createtopic

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"social-network/internal/app"
	"social-network/internal/config"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/validator"

	"github.com/google/uuid"

	topicCommands "social-network/internal/app/topics/commands"
)

const (
	maxUploadSize = 20 << 20 // 20 MB
)

type RequestModel struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	ImagePath   string `json:"imagePath"`
	ImageFile   File   `json:"imageFile"`
	CategoryIDs []int  `json:"categoryIds"`
}
type File struct {
	Content multipart.File
	Header  *multipart.FileHeader
}

type ResponseModel struct {
	UserID  string `json:"userId"`
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

func (h *Handler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	var topicToCreate RequestModel

	err := r.ParseMultipartForm(maxUploadSize) // #nosec G120 -- bounded by maxUploadSize
	if err != nil {
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			"Invalid request payload",
		)
		return
	}
	topicToCreate.Title = r.FormValue("title")
	topicToCreate.Content = r.FormValue("content")
	topicToCreate.CategoryIDs = helpers.ParseStrsToInts(r.Form["categories"])
	topicToCreate.ImageFile.Content, topicToCreate.ImageFile.Header, err = r.FormFile("image_path")
	switch {
	case errors.Is(err, http.ErrMissingFile):
	case err != nil:
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			"Error Processing uploaded file",
		)
		return
	default:
		defer topicToCreate.ImageFile.Content.Close()
		topicToCreate.ImagePath = uuid.New().String() + filepath.Ext(topicToCreate.ImageFile.Header.Filename)
	}

	v := validator.New()

	validator.ValidateCreateTopic(v, &topicToCreate)

	if !v.Valid() {
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			v.ToStringErrors(),
		)

		h.Logger.PrintError(logger.ErrValidationFailed, v.Errors)
		return
	}

	topic, err := h.UserServices.Commands.CreateTopic.Handle(ctx, topicCommands.CreateTopicRequest{
		CategoryIDs: topicToCreate.CategoryIDs,
		Title:       topicToCreate.Title,
		Content:     topicToCreate.Content,
		ImagePath:   topicToCreate.ImagePath,
		ImageFile: topicCommands.TopicImage{
			File:   &topicToCreate.ImageFile.Content,
			Header: topicToCreate.ImageFile.Header,
		},
		User: user,
	})
	if err != nil {
		helpers.RespondWithError(
			w,
			http.StatusInternalServerError,
			"Failed to create topic",
		)

		h.Logger.PrintError(err, nil)

		return
	}

	topicResponse := ResponseModel{
		UserID:  topic.UserID,
		Message: "Topic created successfully",
	}

	helpers.RespondWithJSON(
		w,
		http.StatusCreated,
		nil,
		topicResponse,
	)

	h.Logger.PrintInfo(
		"Topic created successfully",
		map[string]string{
			"user_id": topicResponse.UserID,
		},
	)
}
