package transport

import (
	"errors"
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/topic"
	"social-network/internal/topic/queries"
)

func (h *Handler) GetVoteCounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	topicID, err := helpers.GetQueryInt(r, "id")
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid topic ID")
		return
	}

	var userID *string
	if uid, ok := h.extractUser(r); ok {
		userID = &uid
	}

	vc, err := h.getVotes.Resolve(r.Context(), queries.GetVoteCountsQuery{
		TopicID:     topicID,
		RequesterID: userID,
	})
	if err != nil {
		if errors.Is(err, topic.ErrTopicNotFound) {
			helpers.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := VoteCountsResponse{
		Upvotes:   vc.Upvotes,
		Downvotes: vc.Downvotes,
		Score:     vc.Score,
	}
	helpers.RespondWithJSON(w, http.StatusOK, nil, resp)
}
