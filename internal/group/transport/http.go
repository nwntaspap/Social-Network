package transport

import (
	"context"
	"net/http"

	"social-network/internal/group/commands"
	"social-network/internal/group/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) lookupUser(ctx context.Context, userID string) *UserResult {
	if h.userLookup == nil || userID == "" {
		return nil
	}
	u, err := h.userLookup.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}
	return u
}

func paginatedInfo(total, page, limit int) *helpers.Info {
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}
	return &helpers.Info{
		TotalRecords: total,
		CurrentPage:  page,
		PageSize:     limit,
		TotalPages:   totalPages,
	}
}

func requirePathParam(w http.ResponseWriter, r *http.Request, name, label string) (string, bool) {
	val := r.PathValue(name)
	if val == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, label+" is required")
		return "", false
	}
	return val, true
}

type createGroupBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var body createGroupBody
	if _, err := helpers.ParseBodyRequest(r, &body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	g, err := h.createGroup.Execute(r.Context(), commands.CreateGroupCommand{
		UserID:      userID,
		Title:       body.Title,
		Description: body.Description,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	creator := h.lookupUser(r.Context(), userID)
	membersCount := 1
	helpers.RespondWithJSON(w, http.StatusCreated, nil, toGroupResponse(g, creator, membersCount))
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	pagination := helpers.GetPagination(r)

	res, err := h.listGroups.Resolve(r.Context(), queries.ListGroupsQuery{
		Page: pagination.Page,
		Size: pagination.Limit,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	groups := make([]GroupResponse, 0, len(res.Groups))
	for i := range res.Groups {
		creator := h.lookupUser(r.Context(), res.Groups[i].CreatorID)
		count, _ := h.getGroupMembers.Resolve(r.Context(), queries.GetGroupMembersQuery{
			GroupID: res.Groups[i].ID, Page: 1, Size: 1,
		})
		membersCount := 0
		if count != nil {
			membersCount = count.Total
		}
		groups = append(groups, toGroupResponse(&res.Groups[i], creator, membersCount))
	}

	helpers.RespondWithJSON(w, http.StatusOK, paginatedInfo(res.Total, pagination.Page, pagination.Limit), groups)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
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
	res, err := h.getGroup.Resolve(r.Context(), queries.GetGroupQuery{
		GroupID: groupID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	creator := h.lookupUser(r.Context(), res.Group.CreatorID)
	resp := toGroupDetailResponse(res, creator)
	helpers.RespondWithJSON(w, http.StatusOK, nil, resp)
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if _, err := helpers.ParseBodyRequest(r, &body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	g, err := h.updateGroup.Execute(r.Context(), commands.UpdateGroupCommand{
		GroupID:     groupID,
		UserID:      userID,
		Title:       body.Title,
		Description: body.Description,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	creator := h.lookupUser(r.Context(), g.CreatorID)
	count, _ := h.getGroupMembers.Resolve(r.Context(), queries.GetGroupMembersQuery{
		GroupID: g.ID, Page: 1, Size: 1,
	})
	membersCount := 0
	if count != nil {
		membersCount = count.Total
	}
	helpers.RespondWithJSON(w, http.StatusOK, nil, toGroupResponse(g, creator, membersCount))
}

func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
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

	err := h.deleteGroup.Execute(r.Context(), commands.DeleteGroupCommand{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"status": "ok"})
}

func (h *Handler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
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

	err := h.leaveGroup.Execute(r.Context(), commands.LeaveGroupCommand{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"status": "ok"})
}
