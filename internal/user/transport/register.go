package transport

import (
	"errors"
	"io"
	"net/http"
	"time"

	"social-network/internal/pkg/helpers"
	"social-network/internal/user/commands"
)

const registerMaxUploadSize = 20 << 20 // 20 MB

type registerRequest struct {
	Email       string
	Password    string
	FirstName   string
	LastName    string
	Nickname    string
	DateOfBirth string
	Gender      string
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.PrintError(errors.New("invalid request method"), nil)
		helpers.RespondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}

	if err := r.ParseMultipartForm(registerMaxUploadSize); err != nil { // #nosec G120 -- bounded by registerMaxUploadSize
		h.logger.PrintError(errors.New("invalid form data"), nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid form data")
		return
	}

	req := registerRequest{
		Email:       r.FormValue("email"),
		Password:    r.FormValue("password"),
		FirstName:   r.FormValue("firstName"),
		LastName:    r.FormValue("lastName"),
		Nickname:    r.FormValue("nickname"),
		DateOfBirth: r.FormValue("dateOfBirth"),
		Gender:      r.FormValue("gender"),
	}

	dob, err := time.Parse(time.RFC3339, req.DateOfBirth)
	if err != nil {
		dob, err = time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			helpers.RespondWithError(w, http.StatusBadRequest, "invalid dateOfBirth format, use RFC3339 or YYYY-MM-DD")
			return
		}
	}

	avatarData, avatarFileName, err := readAvatarFile(r)
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(w, http.StatusBadRequest, "failed to read avatar")
		return
	}

	u, err := h.register.Execute(r.Context(), commands.RegisterCommand{
		Email:          req.Email,
		Password:       req.Password,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Nickname:       req.Nickname,
		DateOfBirth:    dob,
		Gender:         req.Gender,
		AvatarData:     avatarData,
		AvatarFileName: avatarFileName,
	})
	if err != nil {
		h.logger.PrintError(err, nil)
		switch {
		case errors.Is(err, commands.ErrEmailTaken),
			errors.Is(err, commands.ErrNicknameTaken):
			helpers.RespondWithError(w, http.StatusConflict, err.Error())
		case errors.Is(err, commands.ErrUnderage),
			errors.Is(err, commands.ErrWeakPassword),
			errors.Is(err, commands.ErrNicknameEmpty),
			errors.Is(err, commands.ErrFirstNameMissing),
			errors.Is(err, commands.ErrLastNameMissing),
			errors.Is(err, commands.ErrInvalidAvatar),
			errors.Is(err, commands.ErrInvalidGender):
			helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			helpers.RespondWithError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	helpers.RespondWithJSON(w, http.StatusCreated, nil, map[string]any{
		"id":    u.ID,
		"email": u.Email,
	})
}

// readAvatarFile extracts the optional "avatar" file from a multipart request.
func readAvatarFile(r *http.Request) ([]byte, string, error) {
	file, header, err := r.FormFile("avatar")
	if err != nil {
		// Field absent or empty is not an error: avatar is optional.
		return nil, "", nil
	}
	defer file.Close()

	buf := make([]byte, registerMaxUploadSize)
	n, readErr := file.Read(buf)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return nil, "", readErr
	}

	name := ""
	if header != nil && header.Filename != "" && n > 0 {
		name = header.Filename
	}
	return buf[:n], name, nil
}
