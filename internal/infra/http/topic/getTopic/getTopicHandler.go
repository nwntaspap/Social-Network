package gettopic

import (
	"context"
	"errors"
	"net/http"

	"social-network/internal/app"
	"social-network/internal/config"
	"social-network/internal/domain/comment"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/infra/storage/sqlite/topics"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/validator"

	topicQueries "social-network/internal/app/topics/queries"
)

type ResponseModel struct {
	UserVote       *int              `json:"userVote"`
	Content        string            `json:"content"`
	ImagePath      string            `json:"imagePath"`
	UserID         string            `json:"userId"`
	OwnerUsername  string            `json:"ownerUsername"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
	Title          string            `json:"title"`
	CategoryNames  []string          `json:"categoryNames"`
	CategoryColors []string          `json:"categoryColors"`
	Comments       []comment.Comment `json:"comments"`
	CategoryIDs    []int             `json:"categoryIds"`
	Upvotes        int               `json:"upvotes"`
	Downvotes      int               `json:"downvotes"`
	Score          int               `json:"score"`
	TopicID        int               `json:"topicId"`
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

func (h *Handler) GetTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.Logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	var userID *string
	user := middleware.GetUserFromContext(r)
	if user != nil {
		userID = &user.ID
	}

	topicID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	val := validator.New()

	testStruct := &struct {
		TopicID int
	}{
		TopicID: topicID,
	}
	validator.ValidateGetTopic(val, testStruct)

	if !val.Valid() {
		h.Logger.PrintError(logger.ErrValidationFailed, val.Errors)
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			val.ToStringErrors(),
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.Config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	topic, err := h.UserServices.Queries.GetTopic.Handle(ctx, topicQueries.GetTopicRequest{
		TopicID: topicID,
		UserID:  userID,
	})
	if err != nil {
		if errors.Is(err, topics.ErrTopicNotFound) {
			helpers.RespondWithError(w, http.StatusNotFound, "Topic not found")
			return
		}

		helpers.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		h.Logger.PrintError(err, nil)
		return
	}

	response := ResponseModel{
		TopicID:        topic.ID,
		CategoryIDs:    topic.CategoryIDs,
		CategoryNames:  topic.CategoryNames,
		CategoryColors: topic.CategoryColors,
		Title:          topic.Title,
		Content:        topic.Content,
		ImagePath:      topic.ImagePath,
		UserID:         topic.UserID,
		OwnerUsername:  topic.OwnerUsername,
		CreatedAt:      topic.CreatedAt,
		UpdatedAt:      topic.UpdatedAt,
		Comments:       topic.Comments,
		Upvotes:        topic.UpvoteCount,
		Downvotes:      topic.DownvoteCount,
		Score:          topic.VoteScore,
		UserVote:       topic.UserVote,
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, response)
}
