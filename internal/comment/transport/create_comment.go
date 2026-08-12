package transport

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"social-network/internal/comment/commands"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
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

	topicID, err := strconv.Atoi(r.FormValue("topicId"))
	if err != nil {
		h.logger.PrintError(errors.New("invalid topic ID"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid topic ID")
		return
	}
	content := strings.TrimSpace(r.FormValue("content"))

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

	c, err := h.createComment.Execute(r.Context(), commands.CreateCommentCommand{
		UserID:        userID,
		TopicID:       topicID,
		Content:       content,
		ImageData:     imageData,
		ImageFileName: imageFileName,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toCommentResponse(c, h.lookupUser(r.Context(), c.UserID))
	helpers.RespondWithJSON(w, http.StatusCreated, nil, resp)
}
