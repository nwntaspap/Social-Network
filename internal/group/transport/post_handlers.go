package transport

import (
	"net/http"

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

	_ = userID
	res, err := h.getGroupFeed.Resolve(r.Context(), queries.GetGroupFeedQuery{
		GroupID: groupID,
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
		cc, _ := h.getGroupPostComments.Resolve(r.Context(), queries.GetGroupPostCommentsQuery{
			PostID: p.ID, Page: 1, Size: 1,
		})
		commentsCount := 0
		if cc != nil {
			commentsCount = cc.Total
		}
		posts = append(posts, toGroupPostResponse(p, user, commentsCount))
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

	var userID string
	if uid, ok := h.extractUser(r); ok {
		userID = uid
	}

	_ = userID
	_, err := h.getGroupChat.Resolve(r.Context(), queries.GetGroupChatQuery{
		GroupID: groupID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, []any{})
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

	title := r.FormValue("title")
	content := r.FormValue("content")

	p, err := h.createGroupPost.Execute(r.Context(), commands.CreateGroupPostCommand{
		GroupID:  groupID,
		AuthorID: userID,
		Title:    title,
		Content:  content,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user := h.lookupUser(r.Context(), p.AuthorID)
	helpers.RespondWithJSON(w, http.StatusCreated, nil, toGroupPostResponse(p, user, 0))
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

	content := r.FormValue("content")

	c, err := h.createGroupPostComment.Execute(r.Context(), commands.CreateGroupPostCommentCommand{
		PostID:   postID,
		AuthorID: userID,
		Content:  content,
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
