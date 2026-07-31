package transport

import (
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/topic/queries"
)

func (h *Handler) GetUserTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	ownerID, err := helpers.GetQueryString(r, "userId")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	requesterID, _ := h.extractUser(r)
	pagination := helpers.GetPagination(r)

	res, err := h.getByUser.Resolve(r.Context(), queries.GetTopicsByUserQuery{
		OwnerID:     ownerID,
		RequesterID: requesterID,
		Page:        pagination.Page,
		Size:        pagination.Limit,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	topics := make([]TopicResponse, 0, len(res.Topics))
	for i := range res.Topics {
		author := h.lookupUser(r.Context(), res.Topics[i].UserID)
		topics = append(topics, toTopicResponse(&res.Topics[i], author))
	}

	totalPages := res.Total / pagination.Limit
	if res.Total%pagination.Limit > 0 {
		totalPages++
	}

	info := &helpers.Info{
		TotalRecords: res.Total,
		CurrentPage:  pagination.Page,
		PageSize:     pagination.Limit,
		TotalPages:   totalPages,
	}
	helpers.RespondWithJSON(w, http.StatusOK, info, topics)
}
