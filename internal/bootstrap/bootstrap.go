package bootstrap

import (
	"os"

	"social-network/internal/app"
	"social-network/internal/app/topics"
	chattransport "social-network/internal/chat/transport"
	commenttransport "social-network/internal/comment/transport"
	"social-network/internal/config"
	coremiddleware "social-network/internal/core/middleware"
	"social-network/internal/core/realtime"
	coresessionstore "social-network/internal/core/session/store"
	"social-network/internal/domain/session"
	eventtransport "social-network/internal/event/transport"
	followtransport "social-network/internal/follow/transport"
	grouptransport "social-network/internal/group/transport"
	"social-network/internal/infra/http/authcookies"
	"social-network/internal/infra/middleware"
	"social-network/internal/infra/realtime/notifications"
	"social-network/internal/infra/storage/sessionstore"
	"social-network/internal/infra/storage/sqlite"
	"social-network/internal/infra/ws"
	oauthtransport "social-network/internal/oauth/transport"
	pkgoauth "social-network/internal/pkg/oAuth"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
	"social-network/internal/platform/logger"
	topictransport "social-network/internal/topic/transport"
	usertransport "social-network/internal/user/transport"

	localstorage "social-network/internal/infra/storage/local"
)

type App struct {
	Services       app.Services
	User           *usertransport.Handler
	Follow         *followtransport.Handler
	Chat           *chattransport.Handler
	Comment        *commenttransport.Handler
	Topic          *topictransport.Handler
	Group          *grouptransport.Handler
	Event          *eventtransport.Handler
	OAuth          *oauthtransport.Handler
	LegacyOAuth    *pkgoauth.OAuth
	Notifier       *notifications.Notifier
	Hub            *ws.Hub
	Realtime       *Realtime
	Middlware      *middleware.Middleware
	SessionManager session.Manager
	CookieManager  *authcookies.Manager
	SessionStore   *coresessionstore.Store
	Logger         logger.Logger
	FileStorage    topics.FileStorageManager
}

func Bootstrap(db database.DB, cfg *config.ServerConfig) *App {
	hub := ws.NewHub()
	rtHub := realtime.NewHub()
	sessionManager := sessionstore.NewSessionManager(db, cfg.SessionManager)
	coreSession := &coreSessionAdapter{inner: sessionManager}
	cookieManager := authcookies.NewManager(cfg.SessionManager)
	coreSessionStore := coresessionstore.NewSessionStore(db, coresessionstore.WithExpiry(cfg.SessionManager.DefaultExpiry))
	sessionCookies := coremiddleware.NewSessionCookies(coremiddleware.CookieConfig{
		Name:     cfg.SessionManager.AccessCookieName,
		Path:     cfg.SessionManager.CookiePath,
		Domain:   cfg.SessionManager.CookieDomain,
		Secure:   cfg.SessionManager.SecureCookie,
		HTTPOnly: cfg.SessionManager.HTTPOnlyCookie,
		SameSite: coremiddleware.ParseSameSite(cfg.SessionManager.SameSite),
	})
	mw := middleware.NewMiddleware(sessionManager, cookieManager)
	repos := sqlite.NewRepositories(db)
	fileStorage := localstorage.NewLocalStorage()
	logger := logger.New(os.Stdout, logger.LevelInfo)
	eventbus, err := eventbus.NewGoBroker()
	if err != nil {
		logger.PrintError(err, nil)
	}

	oauthHandler, legacyOAuth := initOAuth(db, coreSession, sessionCookies, cfg.OAuth, cfg.OAuth.FrontendCallbackURL, logger)

	return &App{
		User:           initUser(db, coreSession, sessionCookies, rtHub.IsOnline, eventbus, logger),
		Follow:         initFollow(db, eventbus, logger),
		Chat:           initChat(db, rtHub, repos.UserRepo, logger),
		Comment:        initComment(db, eventbus, logger),
		Topic:          initTopic(db, eventbus, logger),
		Group:          initGroup(db, eventbus, rtHub.IsOnline, logger),
		Event:          initEvent(db, eventbus, logger),
		OAuth:          oauthHandler,
		LegacyOAuth:    legacyOAuth,
		Hub:            hub,
		Realtime:       initRealtime(db, rtHub, logger),
		Middlware:      mw,
		SessionManager: sessionManager,
		CookieManager:  cookieManager,
		SessionStore:   coreSessionStore,
		Logger:         logger,
		FileStorage:    fileStorage,
	}
}
