package transport

import (
	"errors"
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/topic/queries"
)

func (h *Handler) GetGroupTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	groupID, err := helpers.GetQueryString(r, "groupId")
	if err != nil {
		h.logger.PrintError(errors.New("invalid group ID"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	pagination := helpers.GetPagination(r)

	res, err := h.getByGroup.Resolve(r.Context(), queries.GetTopicsByGroupQuery{
		GroupID: groupID,
		Page:    pagination.Page,
		Size:    pagination.Limit,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	topics := make([]TopicResponse, 0, len(res.Topics))
	for i := range res.Topics {
		author := h.lookupUser(r.Context(), res.Topics[i].UserID)
		topics = append(topics, toTopicResponse(&res.Topics[i], author))
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, paginatedPayload(topics, res.Total, pagination.Page, pagination.Limit))
}
