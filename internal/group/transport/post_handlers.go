package transport

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"social-network/internal/group"
	"social-network/internal/group/commands"
	"social-network/internal/group/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) GetGroupFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	groupID := r.PathValue("groupId")
	if groupID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "groupId is required")
		return
	}

	var userID string
	if uid, ok := h.extractUser(r); ok {
		userID = uid
	}

	pagination := helpers.GetPagination(r)

	res, err := h.getGroupFeed.Resolve(r.Context(), queries.GetGroupFeedQuery{
		GroupID: groupID,
		UserID:  userID,
		Page:    pagination.Page,
		Size:    pagination.Limit,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	posts := make([]GroupPostResponse, 0, len(res.Posts))
	for i := range res.Posts {
		p := &res.Posts[i]
		user := h.lookupUser(r.Context(), p.AuthorID)
		posts = append(posts, toGroupPostResponse(p, user))
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, paginatedPayload(posts, res.Total, pagination.Page, pagination.Limit))
}

func (h *Handler) GetGroupChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	groupID := r.PathValue("groupId")
	if groupID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "groupId is required")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	result, err := h.getGroupChat.Resolve(r.Context(), queries.GetGroupChatQuery{
		GroupID: groupID,
		UserID:  userID,
		Limit:   limit,
	})
	if err != nil {
		if errors.Is(err, group.ErrNotMember) {
			helpers.RespondWithError(w, http.StatusForbidden, "You are not a member of this group")
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, "Failed to get group chat")
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, result.Messages)
}

func (h *Handler) CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	groupID := r.PathValue("groupId")
	if groupID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "groupId is required")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil { // #nosec G120 -- bounded by 20MB limit
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))

	var imageData []byte
	var imageFileName string

	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		buf := make([]byte, 20<<20)
		n, readErr := file.Read(buf)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			helpers.RespondWithError(w, http.StatusBadRequest, "Failed to read image")
			return
		}
		imageData = buf[:n]
		_, header, _ := r.FormFile("image")
		if header != nil {
			imageFileName = filepath.Base(header.Filename)
		}
	}

	p, err := h.createGroupPost.Execute(r.Context(), commands.CreateGroupPostCommand{
		GroupID:       groupID,
		AuthorID:      userID,
		Title:         title,
		Content:       content,
		ImageData:     imageData,
		ImageFileName: imageFileName,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user := h.lookupUser(r.Context(), p.AuthorID)
	helpers.RespondWithJSON(w, http.StatusCreated, nil, toGroupPostResponse(p, user))
}

func (h *Handler) CreateGroupPostComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	postID := r.PathValue("postId")
	if postID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "postId is required")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil { // #nosec G120 -- bounded by 20MB limit
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	content := r.FormValue("content")

	var imageData []byte
	var imageFileName string

	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		buf := make([]byte, 20<<20)
		n, readErr := file.Read(buf)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			helpers.RespondWithError(w, http.StatusBadRequest, "Failed to read image")
			return
		}
		imageData = buf[:n]
		_, header, _ := r.FormFile("image")
		if header != nil {
			imageFileName = filepath.Base(header.Filename)
		}
	}

	c, err := h.createGroupPostComment.Execute(r.Context(), commands.CreateGroupPostCommentCommand{
		PostID:        postID,
		AuthorID:      userID,
		Content:       content,
		ImageData:     imageData,
		ImageFileName: imageFileName,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user := h.lookupUser(r.Context(), c.AuthorID)
	helpers.RespondWithJSON(w, http.StatusCreated, nil, toGroupPostCommentResponse(c, user))
}

func (h *Handler) GetGroupPostComments(w http.ResponseWriter, r *http.Request) {
	postID, ok := requirePathParam(w, r, "postId", "postId")
	if !ok {
		return
	}

	pagination := helpers.GetPagination(r)
	res, err := h.getGroupPostComments.Resolve(r.Context(), queries.GetGroupPostCommentsQuery{
		PostID: postID, Page: pagination.Page, Size: pagination.Limit,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ctx := r.Context()
	comments := make([]GroupPostCommentResponse, 0, len(res.Comments))
	for i := range res.Comments {
		c := &res.Comments[i]
		comments = append(comments, toGroupPostCommentResponse(c, h.lookupUser(ctx, c.AuthorID)))
	}
	helpers.RespondWithJSON(w, http.StatusOK, nil, paginatedPayload(comments, res.Total, pagination.Page, pagination.Limit))
}
