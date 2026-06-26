package bootstrap

import (
	"database/sql"
	"os"
	"time"

	"social-network/internal/app"
	"social-network/internal/app/topics"
	"social-network/internal/config"
	"social-network/internal/domain/session"
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
	Notifier       *notifications.Notifier
	Hub            *ws.Hub
	Middlware      *middleware.Middleware
	SessionManager session.Manager
	CookieManager  *authcookies.Manager
	OAuth          *oauth.OAuth
	Logger         logger.Logger
	FileStorage    topics.FileStorageManager
}

func Bootstrap(db *sql.DB, cfg *config.ServerConfig) *App {
	notifier := notifications.NewNotifier()
	hub := ws.NewHub()
	sessionManager := sessionstore.NewSessionManager(db, cfg.SessionManager)
	cookieManager := authcookies.NewManager(cfg.SessionManager)
	middleware := middleware.NewMiddleware(sessionManager, cookieManager)
	repos := sqlite.NewRepositories(db)
	fileStorage := localstorage.NewLocalStorage()
	services := app.NewServices(repos.UserRepo, repos.CategoryRepo, repos.TopicRepo, repos.CommentRepo, repos.VoteRepo, repos.OauthRepo, repos.ActivityRepo, repos.ChatRepo, repos.NotificationRepo, notifier, hub, fileStorage)
	oAuth := InitOAuth(cfg.OAuth)
	logger := logger.New(os.Stdout, logger.LevelInfo)
	return &App{
		Services:       services,
		Notifier:       notifier,
		Hub:            hub,
		Middlware:      middleware,
		SessionManager: sessionManager,
		CookieManager:  cookieManager,
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
