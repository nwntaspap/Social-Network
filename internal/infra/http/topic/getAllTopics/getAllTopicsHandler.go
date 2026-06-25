package getalltopics

import (
	"context"
	"net/http"
	"strconv"

	"social-network/internal/app"
	"social-network/internal/config"
	"social-network/internal/domain/category"
	"social-network/internal/domain/topic"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/validator"

	topicQueries "social-network/internal/app/topics/queries"
)

type ResponseModel struct {
	Filters    map[string]interface{}   `json:"filters"`
	Topics     []topic.Topic            `json:"topics"`
	Categories []category.Category      `json:"categories"`
	Pagination helpers.PaginationParams `json:"pagination"`
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

func (h *Handler) GetAllTopics(w http.ResponseWriter, r *http.Request) {
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

	params := helpers.NewURLParams(r)

	ctx, cancel := context.WithTimeout(r.Context(), h.Config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	pagination := params.GetPagination()

	orderBy := params.GetQueryStringOr("order_by", "created_at")
	order := params.GetQueryStringOr("order", "desc")
	filter := params.GetQueryStringOr("search", "")
	categoryID := params.GetQueryIntOr("category", 0)

	val := validator.New()

	validator.ValidateGetAllTopics(val, &struct {
		OrderBy    string
		Order      string
		Search     string
		CategoryID int
	}{
		OrderBy:    orderBy,
		Order:      order,
		Search:     filter,
		CategoryID: categoryID,
	})

	if !val.Valid() {
		h.Logger.PrintError(logger.ErrValidationFailed, val.Errors)
		helpers.RespondWithError(
			w,
			http.StatusBadRequest,
			val.ToStringErrors(),
		)
		return
	}
	allTopics, err := h.UserServices.Queries.GetAllTopics.Handle(ctx, topicQueries.GetAllTopicsRequest{
		Page:       pagination.Page,
		Size:       pagination.Limit,
		Offset:     pagination.Offset,
		OrderBy:    orderBy,
		Order:      order,
		Filter:     filter,
		CategoryID: categoryID,
		UserID:     userID,
	})
	if err != nil {
		h.Logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to get topics")
		return
	}

	totalPages := (allTopics.Count + pagination.Limit - 1) / pagination.Limit

	paginationMeta := map[string]interface{}{
		"page":        pagination.Page,
		"limit":       pagination.Limit,
		"total":       allTopics.Count,
		"total_pages": totalPages,
		"has_next":    pagination.Page < totalPages,
		"has_prev":    pagination.Page > 1,
		"next_page":   nil,
		"prev_page":   nil,
	}

	if pagination.Page < totalPages {
		paginationMeta["next_page"] = pagination.Page + 1
	}
	if pagination.Page > 1 {
		paginationMeta["prev_page"] = pagination.Page - 1
	}

	appliedFilters := map[string]interface{}{
		"search":   filter,
		"order_by": orderBy,
		"order":    order,
	}

	response := map[string]interface{}{
		"topics":     allTopics.Topics,
		"pagination": paginationMeta,
		"filters":    appliedFilters,
		"categories": allTopics.Categories,
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, response)

	h.Logger.PrintInfo("Topics retrieved successfully", map[string]string{
		"page":  strconv.Itoa(pagination.Page),
		"count": strconv.Itoa(len(allTopics.Topics)),
		"total": strconv.Itoa(allTopics.Count),
	})
}
