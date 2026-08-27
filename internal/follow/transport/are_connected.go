package transport

import (
	"errors"
	"net/http"

	"social-network/internal/follow/queries"
	"social-network/internal/pkg/helpers"
)

func (h *Handler) AreConnected(w http.ResponseWriter, r *http.Request) {
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

	targetID := r.URL.Query().Get("targetId")
	if targetID == "" {
		h.logger.PrintError(errors.New("targetId is required"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "targetId is required")
		return
	}

	connected, err := h.areConnected.Resolve(r.Context(), queries.AreConnectedQuery{
		UserID:   userID,
		TargetID: targetID,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]bool{"connected": connected})
}
