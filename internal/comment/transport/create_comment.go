package transport

import (
	"net/http"

	"social-network/internal/comment/commands"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		TopicID int    `json:"topicId"`
		Content string `json:"content"`
	}
	if _, err := helpers.ParseBodyRequest(r, &req); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	c, err := h.createComment.Execute(r.Context(), commands.CreateCommentCommand{
		UserID:  userID,
		TopicID: req.TopicID,
		Content: req.Content,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toCommentResponse(c)
	helpers.RespondWithJSON(w, http.StatusCreated, nil, resp)
}
