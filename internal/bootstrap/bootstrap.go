package bootstrap

import (
	"os"

	chattransport "social-network/internal/chat/transport"
	commenttransport "social-network/internal/comment/transport"
	"social-network/internal/config"
	"social-network/internal/core/authcookies"
	coremiddleware "social-network/internal/core/middleware"
	"social-network/internal/core/realtime"
	coresessionstore "social-network/internal/core/session/store"
	eventtransport "social-network/internal/event/transport"
	followtransport "social-network/internal/follow/transport"
	grouptransport "social-network/internal/group/transport"
	oauthtransport "social-network/internal/oauth/transport"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
	"social-network/internal/platform/logger"
	topictransport "social-network/internal/topic/transport"
	usertransport "social-network/internal/user/transport"
)

type App struct {
	User          *usertransport.Handler
	Follow        *followtransport.Handler
	Chat          *chattransport.Handler
	Comment       *commenttransport.Handler
	Topic         *topictransport.Handler
	Group         *grouptransport.Handler
	Event         *eventtransport.Handler
	OAuth         *oauthtransport.Handler
	Realtime      *Realtime
	CookieManager *authcookies.Manager
	SessionStore  *coresessionstore.Store
	Logger        logger.Logger
}

func Bootstrap(db database.DB, cfg *config.ServerConfig) *App {
	rtHub := realtime.NewHub()
	sessionStore := coresessionstore.NewSessionStore(db, coresessionstore.WithExpiry(cfg.SessionManager.DefaultExpiry))
	cookieManager := authcookies.NewManager(cfg.SessionManager)
	sessionCookies := coremiddleware.NewSessionCookies(coremiddleware.CookieConfig{
		Name:     cfg.SessionManager.AccessCookieName,
		Path:     cfg.SessionManager.CookiePath,
		Domain:   cfg.SessionManager.CookieDomain,
		Secure:   cfg.SessionManager.SecureCookie,
		HTTPOnly: cfg.SessionManager.HTTPOnlyCookie,
		SameSite: coremiddleware.ParseSameSite(cfg.SessionManager.SameSite),
	})
	log := logger.New(os.Stdout, logger.LevelInfo)
	eb, err := eventbus.NewGoBroker()
	if err != nil {
		log.PrintError(err, nil)
	}

	oauthHandler := initOAuth(db, sessionStore, sessionCookies, cfg.OAuth, cfg.OAuth.FrontendCallbackURL, log)

	return &App{
		User:          initUser(db, sessionStore, sessionCookies, rtHub.IsOnline, eb, log),
		Follow:        initFollow(db, eb, log),
		Chat:          initChat(db, rtHub, log),
		Comment:       initComment(db, eb, log),
		Topic:         initTopic(db, eb, log),
		Group:         initGroup(db, eb, rtHub.IsOnline, log),
		Event:         initEvent(db, eb, log),
		OAuth:         oauthHandler,
		Realtime:      initRealtime(db, rtHub, log),
		CookieManager: cookieManager,
		SessionStore:  sessionStore,
		Logger:        log,
	}
}
