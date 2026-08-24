package commands

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"social-network/internal/pkg/bcrypt"
	"social-network/internal/pkg/helpers"
	"social-network/internal/pkg/imgutil"
	"social-network/internal/pkg/uuid"
	"social-network/internal/user"
)

var (
	ErrEmailTaken       = errors.New("email already registered")
	ErrUnderage         = errors.New("must be at least 13 years old")
	ErrWeakPassword     = errors.New("password must be at least 8 characters")
	ErrNicknameEmpty    = errors.New("nickname is required")
	ErrNicknameTaken    = errors.New("nickname already taken")
	ErrFirstNameMissing = errors.New("first name is required")
	ErrLastNameMissing  = errors.New("last name is required")
	ErrInvalidAvatar    = errors.New("avatar must be a JPEG, PNG or GIF image")
	ErrInvalidGender    = errors.New("invalid gender value")
)

// allowedGenders mirrors the registration form options.
var allowedGenders = map[string]bool{
	"male":              true,
	"female":            true,
	"other":             true,
	"prefer_not_to_say": true,
}

const uploadsURLPrefix = "/static/images/uploads/"

type RegisterCommand struct {
	Email          string
	Password       string
	FirstName      string
	LastName       string
	Nickname       string
	DateOfBirth    time.Time
	Gender         string
	AvatarData     []byte
	AvatarFileName string
}

type RegisterHandler struct {
	repo   user.Repository
	uuid   uuid.Provider
	crypto bcrypt.Provider
	img    user.ImageStorage
}

func NewRegisterHandler(repo user.Repository, uuid uuid.Provider, crypto bcrypt.Provider, img user.ImageStorage) *RegisterHandler {
	return &RegisterHandler{
		repo:   repo,
		uuid:   uuid,
		crypto: crypto,
		img:    img,
	}
}

func (h *RegisterHandler) Execute(ctx context.Context, cmd RegisterCommand) (*user.User, error) {
	if err := helpers.ValidateEmail(cmd.Email); err != nil {
		return nil, err
	}

	if len(cmd.Password) < 8 {
		return nil, ErrWeakPassword
	}

	if strings.TrimSpace(cmd.FirstName) == "" {
		return nil, ErrFirstNameMissing
	}

	if strings.TrimSpace(cmd.LastName) == "" {
		return nil, ErrLastNameMissing
	}

	if time.Since(cmd.DateOfBirth) < 13*365.25*24*time.Hour {
		return nil, ErrUnderage
	}

	gender := strings.TrimSpace(cmd.Gender)
	if gender == "" {
		gender = "prefer_not_to_say"
	} else if !allowedGenders[gender] {
		return nil, ErrInvalidGender
	}

	nickname := strings.TrimSpace(cmd.Nickname)
	if nickname == "" {
		generated, err := h.generateNickname(ctx, cmd.FirstName)
		if err != nil {
			return nil, err
		}
		nickname = generated
	} else if _, err := h.repo.GetByUsername(ctx, nickname); err == nil {
		return nil, ErrNicknameTaken
	}

	existing, err := h.repo.GetByEmail(ctx, cmd.Email)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	avatarPath := ""
	if len(cmd.AvatarData) > 0 && strings.TrimSpace(cmd.AvatarFileName) != "" {
		if h.img == nil {
			return nil, ErrInvalidAvatar
		}
		mimeType := http.DetectContentType(cmd.AvatarData)
		if !imgutil.AllowedImageTypes[mimeType] {
			return nil, ErrInvalidAvatar
		}
		storedName := h.uuid.NewUUID() + imageExt(mimeType)
		if uploadErr := h.img.Upload(ctx, cmd.AvatarData, storedName); uploadErr != nil {
			return nil, fmt.Errorf("upload avatar: %w", uploadErr)
		}
		avatarPath = uploadsURLPrefix + storedName
	}

	hashed, err := h.crypto.Generate(cmd.Password)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		ID:           h.uuid.NewUUID(),
		Email:        cmd.Email,
		PasswordHash: hashed,
		FirstName:    cmd.FirstName,
		LastName:     cmd.LastName,
		Nickname:     nickname,
		DateOfBirth:  cmd.DateOfBirth,
		Gender:       gender,
		AvatarPath:   avatarPath,
		CreatedAt:    time.Now(),
	}

	if err := h.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

// generateNickname derives "<firstname>_<4 random digits>" and retries until unique.
func (h *RegisterHandler) generateNickname(ctx context.Context, firstName string) (string, error) {
	base := strings.ToLower(strings.TrimSpace(firstName))
	base = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		default:
			return -1
		}
	}, base)
	if base == "" {
		base = "user"
	}
	for range 10 {
		candidate := fmt.Sprintf("%s_%04d", base, rand.Intn(10000)) // #nosec G404 -- non-security identifier
		_, err := h.repo.GetByUsername(ctx, candidate)
		if errors.Is(err, user.ErrUserNotFound) || (err != nil && strings.Contains(err.Error(), user.ErrUserNotFound.Error())) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", ErrNicknameTaken
}

// imageExt maps an allowed MIME type to a safe storage extension.
func imageExt(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}
