package transport

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"social-network/internal/oauth"
	"social-network/internal/oauth/commands"
)

const routesPrefix = "/api/v1/auth/oauth/"

// UserExtractor extracts the authenticated user ID from the request.
type UserExtractor func(r *http.Request) (userID string, ok bool)

// InitiateExecutor starts an OAuth flow.
type InitiateExecutor interface {
	Execute(ctx context.Context, cmd commands.InitiateCommand) (*commands.InitiateResult, error)
}

// CallbackExecutor processes OAuth callbacks.
type CallbackExecutor interface {
	Execute(ctx context.Context, cmd commands.CallbackCommand) (*commands.CallbackResult, error)
}

// Handler holds the OAuth HTTP transport.
type Handler struct {
	initiate     *commands.InitiateHandler
	callbacks    map[string]*commands.CallbackHandler
	extractUser  UserExtractor
	frontendURL  string
	cookieSetter oauth.CookieSetter
}

// NewHandler creates a new OAuth HTTP handler.
// callbacks maps a provider name (e.g. "github", "google") to its callback handler.
func NewHandler(initiate *commands.InitiateHandler, callbacks map[string]*commands.CallbackHandler, cookieSetter oauth.CookieSetter, extractUser UserExtractor, frontendURL string) *Handler {
	return &Handler{
		initiate:     initiate,
		callbacks:    callbacks,
		cookieSetter: cookieSetter,
		extractUser:  extractUser,
		frontendURL:  frontendURL,
	}
}

// RegisterRoutes registers OAuth routes on the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(routesPrefix, h.route)
}

func (h *Handler) route(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, routesPrefix)
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	provider := parts[0]
	action := parts[1]

	switch action {
	case "init":
		h.handleInitiate(w, r, provider, "login")
	case "link":
		h.handleInitiateLink(w, r, provider)
	case "callback":
		h.handleCallback(w, r, provider)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) handleInitiate(w http.ResponseWriter, r *http.Request, provider, flow string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := h.initiate.Execute(r.Context(), commands.InitiateCommand{
		Provider: provider,
		Flow:     flow,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, result.RedirectURL, http.StatusTemporaryRedirect)
}

func (h *Handler) handleInitiateLink(w http.ResponseWriter, r *http.Request, provider string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := h.extractUser(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := h.initiate.Execute(r.Context(), commands.InitiateCommand{
		Provider: provider,
		Flow:     "link",
		UserID:   userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, result.RedirectURL, http.StatusTemporaryRedirect)
}

func (h *Handler) handleCallback(w http.ResponseWriter, r *http.Request, provider string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	callback, ok := h.callbacks[provider]
	if !ok {
		http.NotFound(w, r)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.Error(w, "OAuth error: "+errParam, http.StatusInternalServerError)
		return
	}

	result, err := callback.Execute(r.Context(), commands.CallbackCommand{
		Code:        code,
		State:       state,
		FrontendURL: h.frontendURL,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.Session != nil && h.cookieSetter != nil {
		h.cookieSetter.SetCookies(w, result.Session)
	}

	if result.FrontendURL == "" {
		http.Error(w, "missing redirect target", http.StatusInternalServerError)
		return
	}

	redirectWithHtml(w, result.FrontendURL)
}

// redirectWithHtml sends a 200 HTML page that navigates to target.
// The OAuth callback must not answer with a 3xx redirect: the Next.js rewrite
// proxy that sits in front of the backend on the frontend origin (localhost:3001)
// drops Set-Cookie headers on 3xx responses, so a redirect would lose the session
// cookie. Returning 200 lets the Set-Cookie header ride on a normal response (which
// the proxy forwards), and the script below bounces the browser to the frontend.
func redirectWithHtml(w http.ResponseWriter, target string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Redirecting...</title>
<script>window.location.replace(%q);</script>
</head>
<body>Redirecting...</body>
</html>`, target)
}
