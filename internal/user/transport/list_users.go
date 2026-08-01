package transport

import (
	"net/http"
	"strconv"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user/queries"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if limit < 1 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}

	result, err := h.listUsers.Resolve(r.Context(), queries.ListUsersQuery{
		Query: r.URL.Query().Get("query"),
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	users := make([]map[string]any, 0, len(result.Users))
	for i := range result.Users {
		users = append(users, userResponse(&result.Users[i]))
	}

	totalPages := result.Total / limit
	if result.Total%limit > 0 {
		totalPages++
	}

	helpers.RespondWithJSON(w, http.StatusOK, nil, map[string]any{
		"data":       users,
		"page":       page,
		"pageSize":   limit,
		"totalCount": result.Total,
		"totalPages": totalPages,
	})
}
