package initchat

import (
	"net/http"

	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/pkg/helpers"

	chatcommands "social-network/internal/app/chat/commands"
)

type Handler struct {
	initChatHandler chatcommands.InitChatHandler
	logger          logger.Logger
}

func NewHandler(initChatHandler chatcommands.InitChatHandler, logger logger.Logger) *Handler {
	return &Handler{
		initChatHandler: initChatHandler,
		logger:          logger,
	}
}

type request struct {
	UserID string `json:"user_id"`
}

func (h *Handler) InitChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	me := middleware.GetUserFromContext(r)
	if me == nil {
		helpers.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var body request
	if _, err := helpers.ParseBodyRequest(r, &body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if body.UserID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if body.UserID == me.ID {
		helpers.RespondWithError(w, http.StatusBadRequest, "Cannot open a chat with yourself")
		return
	}

	ctx := r.Context()

	c, err := h.initChatHandler.Handle(ctx, chatcommands.InitChatRequest{MeID: me.ID, UserID: body.UserID})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to initialize chat")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, c)
}
