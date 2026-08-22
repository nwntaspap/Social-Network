package middleware

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Auth struct {
	backendURL string
	client     *http.Client
}

type meResponse struct {
	Data  *meData `json:"data,omitempty"`
	Error string  `json:"error,omitempty"`
}

type meData struct {
	ID string `json:"id,omitempty"`
}

func NewAuth(backendURL string) *Auth {
	return &Auth{
		backendURL: backendURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec
				},
			},
		},
	}
}

func (a *Auth) Required(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, a.backendURL+"/me", nil)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "failed to create auth request")
			return
		}

		for _, c := range r.Cookies() {
			req.AddCookie(c)
		}

		resp, err := a.client.Do(req)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("auth backend unreachable: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			RespondWithError(w, http.StatusUnauthorized, "authentication failed")
			return
		}

		var body meResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			RespondWithError(w, http.StatusUnauthorized, "invalid auth response")
			return
		}
		if body.Data == nil || body.Data.ID == "" {
			RespondWithError(w, http.StatusUnauthorized, "invalid auth response")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, body.Data.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
