package updatetopic

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

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
	Title        string `json:"title"`
	Content      string `json:"content"`
	ImageFile    File   `json:"imageFile"`
	ImagePath    string `json:"imagePath"`
	OldImagePath string
	CategoryIDs  []int `json:"categoryIds"`
	TopicID      int   `json:"topicId"`
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

func (h *Handler) UpdateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	var topicToUpdate RequestModel

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
	topicToUpdate.Title = r.FormValue("title")
	topicToUpdate.OldImagePath = r.FormValue("current_image_path")
	topicToUpdate.Content = r.FormValue("content")
	topicToUpdate.CategoryIDs = helpers.ParseStrsToInts(r.Form["categories"])
	topicToUpdate.TopicID, _ = strconv.Atoi(r.FormValue("topic_id"))
	topicToUpdate.ImageFile.Content, topicToUpdate.ImageFile.Header, err = r.FormFile("image_path")
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
		defer topicToUpdate.ImageFile.Content.Close()
		topicToUpdate.ImagePath = uuid.New().String() + filepath.Ext(topicToUpdate.ImageFile.Header.Filename)
	}

	v := validator.New()

	validator.ValidateCreateTopic(v, &topicToUpdate)

	if !v.Valid() {
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			v.ToStringErrors(),
		)

		h.Logger.PrintError(logger.ErrValidationFailed, v.Errors)
		return
	}

	topic, err := h.UserServices.Commands.UpdateTopic.Handle(ctx, topicCommands.UpdateTopicRequest{
		CategoryIDs:  topicToUpdate.CategoryIDs,
		TopicID:      topicToUpdate.TopicID,
		Title:        topicToUpdate.Title,
		Content:      topicToUpdate.Content,
		ImagePath:    topicToUpdate.ImagePath,
		OldImagePath: topicToUpdate.OldImagePath,
		ImageFile: topicCommands.TopicImage{
			File:   &topicToUpdate.ImageFile.Content,
			Header: topicToUpdate.ImageFile.Header,
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
		Message: "Topic updated successfully",
	}

	helpers.RespondWithJSON(
		w,
		http.StatusCreated,
		nil,
		topicResponse,
	)

	h.Logger.PrintInfo(
		"Topic updated successfully",
		map[string]string{
			"user_id": topicResponse.UserID,
		},
	)
}
