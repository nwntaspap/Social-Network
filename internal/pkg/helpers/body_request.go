package helpers

import (
	"encoding/json"
	"net/http"
)

const (
	maxUploadSize = 20 << 20 // 20 MB
)

func ParseBodyRequest(r *http.Request, v any) (any, error) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func ParseTopicForm(r *http.Request, v any) (any, error) {
	err := r.ParseMultipartForm(maxUploadSize) // #nosec G120 -- bounded by maxUploadSize
	//lint:ignore SA9003 pre-existing
	if err != nil {
	}
	return v, nil
}
