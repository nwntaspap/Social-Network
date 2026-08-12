package transport

import (
	"errors"
	"net/http"

	"social-network/internal/group"
	"social-network/internal/group/queries"
	"social-network/internal/pkg/helpers"
)

// ListMyGroups returns the groups the authenticated user belongs to.
func (h *Handler) ListMyGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	pagination := helpers.GetPagination(r)
	res, err := h.listMyGroups.Resolve(r.Context(), queries.ListMyGroupsQuery{
		UserID: userID,
		Page:   pagination.Page,
		Size:   pagination.Limit,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
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
		groups = append(groups, toGroupResponse(&res.Groups[i], creator, membersCount, res.Groups[i].MembershipStatus))
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, paginatedPayload(groups, res.Total, pagination.Page, pagination.Limit))
}

// GetGroupPresence reports how many of a group's members are currently
// online. Only group members may view it.
func (h *Handler) GetGroupPresence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	groupID := r.PathValue("groupId")
	if groupID == "" {
		h.logger.PrintError(errors.New("groupId is required"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "groupId is required")
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		h.logger.PrintError(errors.New("user not authenticated"), nil)
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	res, err := h.getGroupPresence.Resolve(r.Context(), queries.GetGroupPresenceQuery{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		if errors.Is(err, group.ErrNotMember) {
			helpers.RespondWithError(w, http.StatusForbidden, "You are not a member of this group")
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]any{
		"groupId": res.GroupID,
		"total":   res.Total,
		"online":  res.Online,
		"members": res.Members,
	})
}
