package users

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
)

var (
	ErrDuplicateEmail        = errors.New("email already exists")
	ErrDuplicateUsername     = errors.New("username already exists")
	ErrConstraint            = errors.New("postgres constrain error")
	ErrUnknownConstraint     = errors.New("postgres unknown constraint error")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrUserNotFound          = errors.New("user not found")
	ErrTopicNotFound         = errors.New("topic not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryNotFound      = errors.New("category not found")
)

func MapPQError(err error) error {
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" { // unique_violation
			switch pqErr.Constraint {
			case "users_email_key", "users_email":
				return ErrDuplicateEmail
			case "users_username_key", "users_username":
				return ErrDuplicateUsername
			default:
				return ErrConstraint
			}
		}
		return fmt.Errorf("%w: code=%s constraint=%s", ErrUnknownConstraint, pqErr.Code, pqErr.Constraint)
	}
	return err
}
