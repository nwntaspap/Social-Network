package getme

import (
	"net/http"

	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/pkg/helpers"
)

type Handler struct {
	logger logger.Logger
}

func NewHandler(logger logger.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

type Response struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GetMe handler retrieves the current user from the session in the context.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.PrintError(logger.ErrInvalidRequestMethod, nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	// Get user from context (set by middleware) using the helper function
	user := middleware.GetUserFromContext(r)
	if user == nil {
		helpers.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: User not found")
		return
	}

	response := Response{
		ID:       user.ID,
		Username: user.Nickname,
		Email:    user.Email,
	}
	if user.AvatarURL != nil {
		response.AvatarURL = *user.AvatarURL
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, response)

	h.logger.PrintInfo(
		"User retrieved from session",
		map[string]string{
			"userId": user.ID,
			"name":   user.Nickname,
		},
	)
}
