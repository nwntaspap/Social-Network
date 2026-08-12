package bootstrap

import (
	"context"
	"net/http"
	"time"

	"social-network/internal/config"
	"social-network/internal/core/middleware"
	"social-network/internal/oauth"
	oauthcommands "social-network/internal/oauth/commands"
	oauthstore "social-network/internal/oauth/store"
	oauthtransport "social-network/internal/oauth/transport"
	pkgoauth "social-network/internal/pkg/oAuth"
	"social-network/internal/pkg/oAuth/githubclient"
	"social-network/internal/pkg/oAuth/googleclient"
	"social-network/internal/platform/database"
	"social-network/internal/platform/logger"
)

const stateManagerDefaultLimit = 10

func initOAuth(db database.DB, sessionMgr *coreSessionAdapter, cookies *middleware.SessionCookies, cfg config.OAuthConfig, frontendURL string, logger logger.Logger) (*oauthtransport.Handler, *pkgoauth.OAuth) {
	store := oauthstore.NewSQLiteStore(db)
	sm := pkgoauth.NewStateManager(stateManagerDefaultLimit * time.Minute)

	githubRaw := pkgoauth.Provider(githubclient.NewProvider(
		cfg.GitHub.ClientID,
		cfg.GitHub.ClientSecret,
		cfg.GitHub.RedirectURL,
		cfg.GitHub.Scopes,
	))
	googleRaw := pkgoauth.Provider(googleclient.NewProvider(
		cfg.Google.ClientID,
		cfg.Google.ClientSecret,
		cfg.Google.RedirectURL,
		cfg.Google.TokenURL,
		cfg.Google.Scopes,
	))

	registry := &providerRegistryImpl{
		providers: map[string]oauth.ProviderClient{
			"github": &oauthProviderAdapter{raw: githubRaw},
			"google": &oauthProviderAdapter{raw: googleRaw},
		},
	}

	sc := &sessionCreatorAdapter{sessions: sessionMgr}

	initiateHandler := oauthcommands.NewInitiateHandler(
		&stateManagerAdapter{inner: sm},
		registry,
	)

	callbacks := map[string]*oauthcommands.CallbackHandler{
		"github": oauthcommands.NewCallbackHandler(
			store,
			&stateVerifierAdapter{inner: sm},
			&callbackProviderAdapter{raw: githubRaw, name: "github"},
			sc,
			"github",
		),
		"google": oauthcommands.NewCallbackHandler(
			store,
			&stateVerifierAdapter{inner: sm},
			&callbackProviderAdapter{raw: googleRaw, name: "google"},
			sc,
			"google",
		),
	}

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	legacyOAuth := &pkgoauth.OAuth{
		StateManager:   sm,
		GithubProvider: githubRaw,
		GoogleProvider: googleRaw,
	}

	return oauthtransport.NewHandler(
		initiateHandler,
		callbacks,
		&cookieSetterAdapter{cookies: cookies},
		extractUser,
		frontendURL,
		logger,
	), legacyOAuth
}

type providerRegistryImpl struct {
	providers map[string]oauth.ProviderClient
}

func (r *providerRegistryImpl) Get(name string) (oauth.ProviderClient, bool) {
	p, ok := r.providers[name]
	return p, ok
}

type stateManagerAdapter struct {
	inner *pkgoauth.StateManager
}

func (a *stateManagerAdapter) Generate(data oauthcommands.StateData) (string, error) {
	return a.inner.Generate(pkgoauth.StateData{
		Flow:     data.Flow,
		Provider: data.Provider,
		UserID:   data.UserID,
	})
}

type stateVerifierAdapter struct {
	inner *pkgoauth.StateManager
}

func (a *stateVerifierAdapter) Verify(state string) (oauthcommands.StateData, error) {
	sd, err := a.inner.Verify(state)
	if err != nil {
		return oauthcommands.StateData{}, err
	}
	return oauthcommands.StateData{
		Flow:     sd.Flow,
		Provider: sd.Provider,
		UserID:   sd.UserID,
	}, nil
}

type oauthProviderAdapter struct {
	raw pkgoauth.Provider
}

func (a *oauthProviderAdapter) Name() string                   { return a.raw.Name() }
func (a *oauthProviderAdapter) GetAuthURL(state string) string { return a.raw.GetAuthURL(state) }
func (a *oauthProviderAdapter) ExchangeCode(ctx context.Context, code string) (string, error) {
	return a.raw.ExchangeCode(ctx, code)
}

func (a *oauthProviderAdapter) GetUserInfo(ctx context.Context, accessToken string) (*oauth.ProviderUserInfo, error) {
	info, err := a.raw.GetUserInfo(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	return &oauth.ProviderUserInfo{
		ProviderID: info.ProviderID,
		Email:      info.Email,
		Username:   info.Username,
		Name:       info.Name,
		AvatarURL:  info.AvatarURL,
	}, nil
}

type callbackProviderAdapter struct {
	raw  pkgoauth.Provider
	name string
}

func (a *callbackProviderAdapter) Name() string { return a.raw.Name() }
func (a *callbackProviderAdapter) ExchangeCode(ctx context.Context, code string) (string, error) {
	return a.raw.ExchangeCode(ctx, code)
}

func (a *callbackProviderAdapter) GetUserInfo(ctx context.Context, accessToken string) (*oauth.User, error) {
	info, err := a.raw.GetUserInfo(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	return &oauth.User{
		ProviderID: info.ProviderID,
		Provider:   oauth.Provider(a.name),
		Email:      info.Email,
		Username:   info.Username,
		Name:       info.Name,
		AvatarURL:  info.AvatarURL,
	}, nil
}

type sessionCreatorAdapter struct {
	sessions *coreSessionAdapter
}

func (a *sessionCreatorAdapter) CreateSession(ctx context.Context, userID string) (*oauth.Session, error) {
	sess, err := a.sessions.inner.CreateSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &oauth.Session{
		AccessToken:  sess.AccessToken,
		RefreshToken: sess.RefreshToken,
		ExpiresAt:    sess.Expiry,
	}, nil
}

type cookieSetterAdapter struct {
	cookies *middleware.SessionCookies
}

func (a *cookieSetterAdapter) SetCookies(w http.ResponseWriter, session *oauth.Session) {
	a.cookies.SetAccessCookie(w, session.AccessToken, session.ExpiresAt)
}
