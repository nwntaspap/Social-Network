package transport

import (
	"strings"
	"time"

	"social-network/internal/user"
)

const dateOnlyLayout = "2006-01-02"

// parseDateOnly maps "" to the zero time (clears the field) and otherwise
// requires strict YYYY-MM-DD.
func parseDateOnly(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(dateOnlyLayout, value)
}

func userResponse(u *user.User) map[string]any {
	return map[string]any{
		"id":          u.ID,
		"email":       u.Email,
		"username":    u.Nickname,
		"nickname":    u.Nickname,
		"firstName":   u.FirstName,
		"lastName":    u.LastName,
		"aboutMe":     u.AboutMe,
		"gender":      u.Gender,
		"dateOfBirth": u.DateOfBirth.Format(dateOnlyLayout),
		"avatarUrl":   u.AvatarPath,
		"isPublic":    !u.IsPrivate,
		"createdAt":   u.CreatedAt.Format(time.RFC3339),
	}
}
