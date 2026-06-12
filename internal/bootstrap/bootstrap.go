package bootstrap

import (
	"database/sql"
	"os"
	"time"

	"github.com/arnald/forum/internal/app"
	"github.com/arnald/forum/internal/app/topics"
	"github.com/arnald/forum/internal/config"
	"github.com/arnald/forum/internal/domain/session"
	"github.com/arnald/forum/internal/infra/http/authcookies"
	"github.com/arnald/forum/internal/infra/logger"
	"github.com/arnald/forum/internal/infra/middleware"
	"github.com/arnald/forum/internal/infra/realtime/notifications"
	localstorage "github.com/arnald/forum/internal/infra/storage/local"
	"github.com/arnald/forum/internal/infra/storage/postgres"
	"github.com/arnald/forum/internal/infra/storage/postgres/sessionstore"
	"github.com/arnald/forum/internal/infra/ws"
	oauth "github.com/arnald/forum/internal/pkg/oAuth"
	"github.com/arnald/forum/internal/pkg/oAuth/githubclient"
	"github.com/arnald/forum/internal/pkg/oAuth/googleclient"
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
	repos := postgres.NewRepositories(db)
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
