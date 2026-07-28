package transport

import (
	"net/http"

	"social-network/internal/group/commands"
	"social-network/internal/group/queries"
	"social-network/internal/pkg/helpers"
)

type inviteBody struct {
	UserID string `json:"userId"`
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
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

	var body inviteBody
	if _, err := helpers.ParseBodyRequest(r, &body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if body.UserID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "userId is required")
		return
	}

	inv, err := h.inviteMember.Execute(r.Context(), commands.InviteMemberCommand{
		GroupID:   groupID,
		InviterID: userID,
		InviteeID: body.UserID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	inviter := h.lookupUser(r.Context(), inv.InviterID)
	invitee := h.lookupUser(r.Context(), inv.InviteeID)
	helpers.RespondWithJSON(w, http.StatusCreated, nil, toInvitationResponse(inv, inviter, invitee))
}

func (h *Handler) RespondInvite(w http.ResponseWriter, r *http.Request) {
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
		Action string `json:"action"`
	}
	if _, err := helpers.ParseBodyRequest(r, &body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	accept := body.Action == "accept"

	err := h.respondInvite.Execute(r.Context(), commands.RespondInviteCommand{
		GroupID:   groupID,
		InviteeID: userID,
		Accept:    accept,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"status": "ok"})
}

func (h *Handler) RequestJoin(w http.ResponseWriter, r *http.Request) {
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

	jr, err := h.requestJoin.Execute(r.Context(), commands.RequestJoinCommand{
		GroupID:     groupID,
		RequesterID: userID,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	requester := h.lookupUser(r.Context(), jr.RequesterID)
	g, _ := h.getGroup.Resolve(r.Context(), queries.GetGroupQuery{GroupID: jr.GroupID})
	var groupBrief *GroupBrief
	if g != nil {
		groupBrief = &GroupBrief{ID: g.Group.ID, Title: g.Group.Title}
	}
	helpers.RespondWithJSON(w, http.StatusCreated, nil, toJoinRequestResponse(jr, groupBrief, requester))
}

type respondJoinBody struct {
	Action string `json:"action"`
}

func (h *Handler) RespondJoin(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	requestID := r.PathValue("requestId")
	if requestID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "requestId is required")
		return
	}

	var body respondJoinBody
	if _, err := helpers.ParseBodyRequest(r, &body); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	accept := body.Action == "accept"

	err := h.respondJoin.Execute(r.Context(), commands.RespondJoinCommand{
		RequestID: requestID,
		AdminID:   userID,
		Accept:    accept,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]string{"status": "ok"})
}

func (h *Handler) GetPendingJoinRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	_, ok := h.extractUser(r)
	if !ok {
		helpers.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	groupID := r.PathValue("groupId")
	if groupID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "groupId is required")
		return
	}

	_ = groupID

	helpers.RespondWithJSON(w, http.StatusOK, nil, []any{})
}

func (h *Handler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("groupId")
	if groupID == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "groupId is required")
		return
	}

	pagination := helpers.GetPagination(r)
	res, err := h.getGroupMembers.Resolve(r.Context(), queries.GetGroupMembersQuery{
		GroupID: groupID, Page: pagination.Page, Size: pagination.Limit,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ctx := r.Context()
	members := make([]GroupMemberResponse, 0, len(res.Members))
	for i := range res.Members {
		m := &res.Members[i]
		members = append(members, toGroupMemberResponse(m, h.lookupUser(ctx, m.UserID)))
	}
	helpers.RespondWithJSON(w, http.StatusOK, paginatedInfo(res.Total, pagination.Page, pagination.Limit), members)
}
