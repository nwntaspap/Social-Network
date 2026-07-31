package transport

import (
	"time"

	"social-network/internal/user"
)

const dateOnlyLayout = "2006-01-02"

func userResponse(u *user.User) map[string]any {
	return map[string]any{
		"id":          u.ID,
		"email":       u.Email,
		"username":    u.Nickname,
		"nickname":    u.Nickname,
		"firstName":   u.FirstName,
		"lastName":    u.LastName,
		"aboutMe":     u.AboutMe,
		"dateOfBirth": u.DateOfBirth.Format(dateOnlyLayout),
		"avatarUrl":   u.AvatarPath,
		"isPublic":    !u.IsPrivate,
		"createdAt":   u.CreatedAt.Format(time.RFC3339),
	}
}
