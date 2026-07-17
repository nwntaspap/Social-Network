package transport

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"social-network/internal/pkg/helpers"
	"social-network/internal/topic"
	"social-network/internal/topic/commands"
)

func (h *Handler) UpdateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	topicID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid topic ID")
		return
	}

	//nolint:gosec // max 20MB multipart upload
	if parseErr := r.ParseMultipartForm(20 << 20); parseErr != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	visibilityStr := r.FormValue("privacy")

	var visibility topic.Visibility
	switch visibilityStr {
	case "followers":
		visibility = topic.VisibilityFollowers
	case "private":
		visibility = topic.VisibilityPrivate
	default:
		visibility = topic.VisibilityPublic
	}

	var imageData []byte
	var imageFileName string

	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		buf := make([]byte, 20<<20)
		n, readErr := file.Read(buf)
		if readErr != nil {
			helpers.RespondWithError(w, http.StatusBadRequest, "Failed to read image")
			return
		}
		imageData = buf[:n]
		_, header, _ := r.FormFile("image")
		if header != nil {
			imageFileName = filepath.Base(header.Filename)
		}
	}

	var allowedUserIDs []string
	if allowedStr := r.FormValue("allowedUserIds"); allowedStr != "" {
		if unmarshalErr := json.Unmarshal([]byte(allowedStr), &allowedUserIDs); unmarshalErr != nil {
			helpers.RespondWithError(w, http.StatusBadRequest, "Invalid allowedUserIds")
			return
		}
	}

	cmd := commands.UpdateTopicCommand{
		TopicID:        topicID,
		UserID:         userID,
		Title:          title,
		Content:        content,
		ImageData:      imageData,
		ImageFileName:  imageFileName,
		Visibility:     visibility,
		AllowedUserIDs: allowedUserIDs,
	}

	top, err := h.updateTopic.Execute(r.Context(), cmd)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toTopicResponse(top)
	helpers.RespondWithJSON(w, http.StatusOK, nil, resp)
}
