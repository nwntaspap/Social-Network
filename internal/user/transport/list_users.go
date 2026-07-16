package transport

import (
	"net/http"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user/queries"
)

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	result, err := h.listUsers.Resolve(r.Context(), queries.ListUsersQuery{})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	users := make([]map[string]any, 0, len(result.Users))
	for _, u := range result.Users {
		users = append(users, map[string]any{
			"id":        u.ID,
			"email":     u.Email,
			"firstName": u.FirstName,
			"lastName":  u.LastName,
			"nickname":  u.Nickname,
		})
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, users)
}
