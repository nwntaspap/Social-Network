package transport

import (
	"context"
	"net/http"
	"strings"

	"social-network/internal/oauth/commands"
)

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
	initiate    *commands.InitiateHandler
	callback    *commands.CallbackHandler
	extractUser UserExtractor
	frontendURL string
}

// NewHandler creates a new OAuth HTTP handler.
func NewHandler(initiate *commands.InitiateHandler, callback *commands.CallbackHandler, extractUser UserExtractor, frontendURL string) *Handler {
	return &Handler{
		initiate:    initiate,
		callback:    callback,
		extractUser: extractUser,
		frontendURL: frontendURL,
	}
}

// RegisterRoutes registers OAuth routes on the provided mux.
// NOTE: Not wired into bootstrap until S5-BE-83.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/oauth/", h.route)
}

func (h *Handler) route(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/auth/oauth/")
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

func (h *Handler) handleCallback(w http.ResponseWriter, r *http.Request, _ string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.Error(w, "OAuth error: "+errParam, http.StatusInternalServerError)
		return
	}

	result, err := h.callback.Execute(r.Context(), commands.CallbackCommand{
		Code:        code,
		State:       state,
		FrontendURL: h.frontendURL,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.Session != nil {
		// Session cookie setting will be handled by the cookie adapter in bootstrap
		// For now, redirect to frontend with success params
		http.Redirect(w, r, result.FrontendURL, http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, result.FrontendURL, http.StatusTemporaryRedirect)
}
