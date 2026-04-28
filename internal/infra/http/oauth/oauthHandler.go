package oauthlogin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	oauthservice "github.com/arnald/forum/internal/app/oauth"
	"github.com/arnald/forum/internal/config"
	"github.com/arnald/forum/internal/domain/session"
	"github.com/arnald/forum/internal/infra/logger"
	"github.com/arnald/forum/internal/pkg/helpers"
	oauthpkg "github.com/arnald/forum/internal/pkg/oAuth"
)

type OAuthHandler struct {
	provider       oauthpkg.Provider
	config         *config.ServerConfig
	loginService   *oauthservice.OAuthService
	stateManager   *oauthpkg.StateManager
	sessionManager session.Manager
	logger         logger.Logger
}

func NewOAuthHandler(
	provider oauthpkg.Provider,
	config *config.ServerConfig,
	loginService *oauthservice.OAuthService,
	stateManager *oauthpkg.StateManager,
	sessionManager session.Manager,
	logger logger.Logger,
) *OAuthHandler {
	return &OAuthHandler{
		provider:       provider,
		config:         config,
		loginService:   loginService,
		stateManager:   stateManager,
		sessionManager: sessionManager,
		logger:         logger,
	}
}

func (h *OAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := h.stateManager.Generate()
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
	log.Println("hello")
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
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

	err := h.stateManager.Verify(state)
	if err != nil {
		h.logger.PrintError(err, nil)
		http.Error(
			w,
			"problem with oauth STATE, SEE LOGGER",
			http.StatusInternalServerError,
		)
		return
	}

	user, err := h.loginService.Login(
		ctx,
		code,
		h.provider,
	)
	if err != nil {
		h.logger.PrintError(err, map[string]string{
			"action":   "oauth_login",
			"provider": h.provider.Name(),
		})
		// Respond with a popup bridge so SPA popups can receive the error
		frontendCallbackBase := h.config.OAuth.FrontendCallbackURL
		switch h.provider.Name() {
		case "github":
			frontendCallbackBase = h.config.OAuth.GitHub.FrontendCallbackURL
		case "google":
			frontendCallbackBase = h.config.OAuth.Google.FrontendCallbackURL
		}
		frontendOrigin := frontendCallbackBase
		if u, perr := url.Parse(frontendCallbackBase); perr == nil {
			frontendOrigin = u.Scheme + "://" + u.Host
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		errBridge := fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"></head><body><script>
try { window.opener.postMessage({ type: 'oauth', status: 'error', message: %q, provider: %q }, %q); } catch(e) {}
window.close();
</script></body></html>`, err.Error(), h.provider.Name(), frontendOrigin)
		_, _ = w.Write([]byte(errBridge))
		return
	}

	session, err := h.sessionManager.CreateSession(r.Context(), user.ID)
	if err != nil {
		h.logger.PrintError(err, nil)
		// Respond with a bridge error page so popup can know
		frontendCallbackBase := h.config.OAuth.FrontendCallbackURL
		switch h.provider.Name() {
		case "github":
			frontendCallbackBase = h.config.OAuth.GitHub.FrontendCallbackURL
		case "google":
			frontendCallbackBase = h.config.OAuth.Google.FrontendCallbackURL
		}
		frontendOrigin := frontendCallbackBase
		if u, perr := url.Parse(frontendCallbackBase); perr == nil {
			frontendOrigin = u.Scheme + "://" + u.Host
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		errBridge := fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"></head><body><script>
try { window.opener.postMessage({ type: 'oauth', status: 'error', message: %q, provider: %q }, %q); } catch(e) {}
window.close();
</script></body></html>`, "error at creating session", h.provider.Name(), frontendOrigin)
		_, _ = w.Write([]byte(errBridge))
		return
	}

	// Set secure HttpOnly cookies so the browser stores session tokens automatically.
	// This avoids exposing tokens in URLs. Make sure SetCookies is called before
	// writing the response so Set-Cookie headers are included.
	h.sessionManager.SetCookies(w, session)

	// Determine the provider-specific frontend callback URL (used only to derive origin)
	var frontendCallbackBase string
	switch h.provider.Name() {
	case "github":
		frontendCallbackBase = h.config.OAuth.GitHub.FrontendCallbackURL
	case "google":
		frontendCallbackBase = h.config.OAuth.Google.FrontendCallbackURL
	default:
		frontendCallbackBase = h.config.OAuth.FrontendCallbackURL
	}
	frontendOrigin := frontendCallbackBase
	if u, perr := url.Parse(frontendCallbackBase); perr == nil {
		frontendOrigin = u.Scheme + "://" + u.Host
	}

	// Respond with a small HTML bridge that posts a message to the opener window
	// and then closes the popup. The SPA should listen for this message and then
	// call /api/v1/me (with credentials) to fetch the logged-in user.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	bridge := fmt.Sprintf(`<!doctype html>
<html>
  <head><meta charset="utf-8"></head>
  <body>
    <script>
      (function() {
        try {
          var payload = { type: 'oauth', status: 'success', provider: %q };
          window.opener.postMessage(payload, %q);
        } catch (e) {}
        window.close();
      })();
    </script>
  </body>
</html>`, h.provider.Name(), frontendOrigin)
	_, _ = w.Write([]byte(bridge))

	h.logger.PrintInfo(
		"User logged in via "+h.provider.Name(),
		map[string]string{
			"user_id":  user.ID,
			"username": user.Nickname,
			"provider": h.provider.Name(),
		})
}
