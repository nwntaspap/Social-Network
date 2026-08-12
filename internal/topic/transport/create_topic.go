package transport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"social-network/internal/pkg/helpers"
	"social-network/internal/topic"
	"social-network/internal/topic/commands"
)

func (h *Handler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		h.logger.PrintError(errors.New("user not authenticated"), nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil { // #nosec G120 -- bounded by 20MB limit
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	visibilityStr := r.FormValue("privacy")
	groupID := r.FormValue("groupId")

	if title == "" {
		h.logger.PrintError(errors.New("title is required"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if content == "" {
		h.logger.PrintError(errors.New("content is required"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Content is required")
		return
	}

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
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			h.logger.PrintError(readErr, nil)
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
			h.logger.PrintError(errors.New("invalid allowedUserIds"), nil)
			helpers.RespondWithError(w, http.StatusBadRequest, "Invalid allowedUserIds")
			return
		}
	}

	cmd := commands.CreateTopicCommand{
		UserID:         userID,
		Title:          title,
		Content:        content,
		ImageData:      imageData,
		ImageFileName:  imageFileName,
		Visibility:     visibility,
		AllowedUserIDs: allowedUserIDs,
	}
	if groupID != "" {
		cmd.GroupID = &groupID
	}

	top, err := h.createTopic.Execute(r.Context(), cmd)
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toTopicResponse(top, h.lookupUser(r.Context(), userID))
	helpers.RespondWithJSON(w, http.StatusCreated, nil, resp)
}
