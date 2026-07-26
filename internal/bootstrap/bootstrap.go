package bootstrap

import (
	"database/sql"
	"os"
	"time"

	"social-network/internal/app"
	"social-network/internal/app/topics"
	chattransport "social-network/internal/chat/transport"
	"social-network/internal/config"
	coresessionstore "social-network/internal/core/session/store"
	"social-network/internal/domain/session"
	followtransport "social-network/internal/follow/transport"
	grouptransport "social-network/internal/group/transport"
	"social-network/internal/infra/http/authcookies"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/infra/realtime/notifications"
	"social-network/internal/infra/storage/sessionstore"
	"social-network/internal/infra/storage/sqlite"
	"social-network/internal/infra/ws"
	"social-network/internal/pkg/oAuth/githubclient"
	"social-network/internal/pkg/oAuth/googleclient"

	localstorage "social-network/internal/infra/storage/local"

	oauth "social-network/internal/pkg/oAuth"
)

const stateManagerDefaultLimit = 10

type App struct {
	Services       app.Services
	Follow         *followtransport.Handler
	Chat           *chattransport.Handler
	Group          *grouptransport.Handler
	Notifier       *notifications.Notifier
	Hub            *ws.Hub
	Middlware      *middleware.Middleware
	SessionManager session.Manager
	CookieManager  *authcookies.Manager
	SessionStore   *coresessionstore.Store
	OAuth          *oauth.OAuth
	Logger         logger.Logger
	FileStorage    topics.FileStorageManager
}

func Bootstrap(db *sql.DB, cfg *config.ServerConfig) *App {
	notifier := notifications.NewNotifier()
	hub := ws.NewHub()
	sessionManager := sessionstore.NewSessionManager(db, cfg.SessionManager)
	cookieManager := authcookies.NewManager(cfg.SessionManager)
	coreSessionStore := coresessionstore.NewSessionStore(db, coresessionstore.WithExpiry(cfg.SessionManager.DefaultExpiry))
	middleware := middleware.NewMiddleware(sessionManager, cookieManager)
	repos := sqlite.NewRepositories(db)
	fileStorage := localstorage.NewLocalStorage()
	services := app.NewServices(repos.UserRepo, repos.CategoryRepo, repos.TopicRepo, repos.CommentRepo, repos.VoteRepo, repos.OauthRepo, repos.ActivityRepo, repos.ChatRepo, repos.NotificationRepo, notifier, hub, fileStorage)
	oAuth := InitOAuth(cfg.OAuth)
	logger := logger.New(os.Stdout, logger.LevelInfo)
	return &App{
		Services:       services,
		Follow:         initFollow(db),
		Chat:           initChat(db, hub, repos.UserRepo),
		Group:          initGroup(db),
		Notifier:       notifier,
		Hub:            hub,
		Middlware:      middleware,
		SessionManager: sessionManager,
		CookieManager:  cookieManager,
		SessionStore:   coreSessionStore,
		OAuth:          oAuth,
		Logger:         logger,
		FileStorage:    fileStorage,
	}
}

func InitOAuth(cfg config.OAuthConfig) *oauth.OAuth {
	return &oauth.OAuth{
		StateManager: oauth.NewStateManager(stateManagerDefaultLimit * time.Minute),
		GithubProvider: githubclient.NewProvider(
			cfg.GitHub.ClientID,
			cfg.GitHub.ClientSecret,
			cfg.GitHub.RedirectURL,
			cfg.GitHub.Scopes,
		),
		GoogleProvider: googleclient.NewProvider(
			cfg.Google.ClientID,
			cfg.Google.ClientSecret,
			cfg.Google.RedirectURL,
			cfg.Google.TokenURL,
			cfg.Google.Scopes,
		),
	}
}
