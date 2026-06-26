package oauthlogin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"social-network/internal/config"
	"social-network/internal/domain/oauth"
	"social-network/internal/domain/session"
	"social-network/internal/infra/http/authcookies"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/pkg/helpers"

	oauthservice "social-network/internal/app/oauth"

	oauthpkg "social-network/internal/pkg/oAuth"
)

type OAuthHandler struct {
	provider       oauthpkg.Provider
	config         *config.ServerConfig
	loginService   *oauthservice.OAuthService
	stateManager   *oauthpkg.StateManager
	sessionManager session.Manager
	cookieManager  *authcookies.Manager
	logger         logger.Logger
}

func NewOAuthHandler(
	provider oauthpkg.Provider,
	config *config.ServerConfig,
	loginService *oauthservice.OAuthService,
	stateManager *oauthpkg.StateManager,
	sessionManager session.Manager,
	logger logger.Logger,
	cookieManager *authcookies.Manager,
) *OAuthHandler {
	return &OAuthHandler{
		provider:       provider,
		config:         config,
		loginService:   loginService,
		stateManager:   stateManager,
		sessionManager: sessionManager,
		logger:         logger,
		cookieManager:  cookieManager,
	}
}

func (h *OAuthHandler) Link(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)

	var stateData oauthpkg.StateData
	stateData.Flow = "link"
	stateData.Provider = h.provider.Name()
	fmt.Println(user.ID)
	stateData.UserID = user.ID

	state, err := h.stateManager.Generate(stateData)
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	authURL := h.provider.GetAuthURL(state)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var stateData oauthpkg.StateData
	stateData.Flow = "login"
	stateData.Provider = h.provider.Name()
	state, err := h.stateManager.Generate(stateData)
	if err != nil {
		h.logger.PrintError(err, nil)
		helpers.RespondWithError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	authURL := h.provider.GetAuthURL(state)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}
	params := url.Values{}
	var frontendCallbackBase string
	switch h.provider.Name() {
	case "github":
		frontendCallbackBase = h.config.OAuth.GitHub.FrontendCallbackURL
	case "google":
		frontendCallbackBase = h.config.OAuth.Google.FrontendCallbackURL
	default:
		frontendCallbackBase = h.config.OAuth.FrontendCallbackURL
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.config.Timeouts.HandlerTimeouts.UserRegister)
	defer cancel()

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	errParam := r.URL.Query().Get("error")
	if errParam != "" {
		h.logger.PrintError(fmt.Errorf("%w: %s", ErrInParameters, errParam), nil)
		http.Error(
			w,
			"problem with oatuh, see logger",
			http.StatusInternalServerError,
		)
		return
	}

	if code == "" {
		h.logger.PrintError(ErrCodeMissing, nil)
		http.Error(
			w,
			"no code in callback",
			http.StatusInternalServerError,
		)
		return
	}

	stateData, err := h.stateManager.Verify(state)
	if err != nil {
		h.logger.PrintError(err, nil)
		http.Error(
			w,
			"problem with oauth STATE, SEE LOGGER",
			http.StatusInternalServerError,
		)
		return
	}

	switch stateData.Flow {
	case "login":
		user, err := h.loginService.Login(
			ctx,
			code,
			h.provider,
		)

		if errors.Is(err, oauth.ErrUserWithEmailExists) {
			h.logger.PrintError(err, map[string]string{
				"action":   "oauth_login",
				"provider": h.provider.Name(),
			})
			params.Add("flow", "login")
			params.Add("error", "email_exists")
			params.Add("provider", h.provider.Name())
			frontendCallbackURL := fmt.Sprintf("%s?%s", frontendCallbackBase, params.Encode())
			http.Redirect(w, r, frontendCallbackURL, http.StatusTemporaryRedirect)
			return
		}
		if err != nil {
			h.logger.PrintError(err, map[string]string{
				"action":   "oauth_login",
				"provider": h.provider.Name(),
			})
			params.Add("flow", "login")
			params.Add("error", "errorAtOauthLogin")
			params.Add("provider", h.provider.Name())
			frontendCallbackURL := fmt.Sprintf("%s?%s", frontendCallbackBase, params.Encode())
			http.Redirect(w, r, frontendCallbackURL, http.StatusTemporaryRedirect)
			return
		}

		session, err := h.sessionManager.CreateSession(r.Context(), user.ID)
		if err != nil {
			h.logger.PrintError(err, nil)
			params.Add("flow", "login")
			params.Add("error", "errorCreatingSession")
			params.Add("provider", h.provider.Name())
			frontendCallbackURL := fmt.Sprintf("%s?%s", frontendCallbackBase, params.Encode())
			http.Redirect(w, r, frontendCallbackURL, http.StatusTemporaryRedirect)
			return
		}

		h.cookieManager.SetCookies(w, session)
		params.Add("flow", "login")
		params.Add("success", "ok")
		params.Add("provider", h.provider.Name())
		frontendCallbackURL := fmt.Sprintf("%s?%s", frontendCallbackBase, params.Encode())
		http.Redirect(w, r, frontendCallbackURL, http.StatusTemporaryRedirect)

		h.logger.PrintInfo(
			"User logged in via "+h.provider.Name(),
			map[string]string{
				"user_id":  user.ID,
				"username": user.Nickname,
				"provider": h.provider.Name(),
			},
		)
	case "link":
		err := h.loginService.Link(
			ctx,
			stateData.UserID,
			code,
			h.provider,
		)

		if errors.Is(err, oauth.ErrProviderAccountBelongsToAnotherUser) {
			h.logger.PrintError(err, map[string]string{
				"action":   "oauth_link",
				"provider": h.provider.Name(),
			})
			params.Add("flow", "link")
			params.Add("error", "providerAccountBelongsToAnotherUser")
			params.Add("provider", h.provider.Name())
			frontendCallbackURL := fmt.Sprintf("%s?%s", frontendCallbackBase, params.Encode())
			http.Redirect(w, r, frontendCallbackURL, http.StatusTemporaryRedirect)
			return
		}
		if errors.Is(err, oauth.ErrAlreadyLinkedToProvider) {
			h.logger.PrintError(err, map[string]string{
				"action":   "oauth_link",
				"provider": h.provider.Name(),
			})
			params.Add("flow", "link")
			params.Add("error", "alreadyLinkedToProvider")
			params.Add("provider", h.provider.Name())
			frontendCallbackURL := fmt.Sprintf("%s?%s", frontendCallbackBase, params.Encode())
			http.Redirect(w, r, frontendCallbackURL, http.StatusTemporaryRedirect)
			return
		}
		if err != nil {
			h.logger.PrintError(err, map[string]string{
				"action":   "oauth_login",
				"provider": h.provider.Name(),
			})
			params.Add("flow", "link")
			params.Add("error", "errorAtOauthLogin")
			params.Add("provider", h.provider.Name())
			frontendCallbackURL := fmt.Sprintf("%s?%s", frontendCallbackBase, params.Encode())
			http.Redirect(w, r, frontendCallbackURL, http.StatusTemporaryRedirect)
			return
		}

		params.Add("flow", "link")
		params.Add("success", "ok")
		params.Add("provider", h.provider.Name())
		frontendCallbackURL := fmt.Sprintf("%s?%s", frontendCallbackBase, params.Encode())
		http.Redirect(w, r, frontendCallbackURL, http.StatusTemporaryRedirect)

		h.logger.PrintInfo(
			"User Acount linked to provider"+h.provider.Name(),
			map[string]string{
				"user_id":  stateData.UserID,
				"provider": h.provider.Name(),
			},
		)
	default:
		params.Add("error", "flowNotRecognised")
		frontendCallbackURL := fmt.Sprintf("%s?%s", frontendCallbackBase, params.Encode())
		http.Redirect(w, r, frontendCallbackURL, http.StatusTemporaryRedirect)
	}
}
