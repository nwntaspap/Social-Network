package imgutil

import (
	"errors"
	"net/http"
)

var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

func ValidateImageHeader(data []byte) error {
	mime := http.DetectContentType(data)
	if !AllowedImageTypes[mime] {
		return errors.New("invalid image format")
	}
	return nil
}
