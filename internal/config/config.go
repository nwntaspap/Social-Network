package config

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/arnald/forum/internal/pkg/helpers"
	"github.com/arnald/forum/internal/pkg/path"
)

const (
	readTimeout                     = 5
	writeTimeout                    = 10
	idleTimeout                     = 15
	configParts                     = 2
	defaultExpiry                   = 86400
	cleanupInternal                 = 3600
	maxSessionsPerUser              = 5
	sessionIDLenght                 = 32
	userRegisterTimeout             = 15
	refreshTokenExpiry              = 30
	userLoginTimeout                = 15
	defaultRateLimitCleanupSeconds  = 60
	defaultRateLimitWindowSeconds   = 60
	defaultRateLimitRequestCapacity = 100
)

var (
	ErrMissingServerHost    = errors.New("missing SERVER_HOST in config")
	ErrServerPortNotInteger = errors.New("invalid SERVER_PORT: must be integer")
)

type ServerConfig struct {
	OAuth          OAuthConfig
	Host           string
	Port           string
	Environment    string
	APIContext     string
	TLSCertFile    string
	TLSKeyFile     string
	Database       DatabaseConfig
	SessionManager SessionManagerConfig
	Timeouts       TimeoutsConfig
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	RateLimit      RateLimitConfig
}

type RateLimitConfig struct {
	Enabled       bool
	RequestsLimit int
	WindowSeconds int64
	Cleanup       time.Duration
}

type OAuthConfig struct {
	FrontendCallbackURL string
	GitHub              GitHubOAuthConfig
	Google              GoogleOAuthConfig
}

type GitHubOAuthConfig struct {
	ClientID            string
	ClientSecret        string
	RedirectURL         string
	FrontendCallbackURL string
	Scopes              []string
}

type GoogleOAuthConfig struct {
	ClientID            string
	ClientSecret        string
	RedirectURL         string
	FrontendCallbackURL string
	TokenURL            string
	Scopes              []string
}
type DatabaseConfig struct {
	Driver              string
	PostgresURL         string
	Path                string
	Pragma_Foreign_Keys string
	Pragma_Journal_Mode string
	MigrateOnStart      bool
	SeedOnStart         bool
	OpenConn            int
}

type SessionManagerConfig struct {
	AccessCookieName   string
	RefreshCookieName  string
	CookiePath         string
	CookieDomain       string
	SameSite           string
	DefaultExpiry      time.Duration
	CleanupInterval    time.Duration
	MaxSessionsPerUser int
	SessionIDLength    int
	SecureCookie       bool
	HTTPOnlyCookie     bool
	EnablePersistence  bool
	LogSessions        bool
	RefreshTokenExpiry time.Duration
}

type TimeoutsConfig struct {
	HandlerTimeouts  HandlerTimeoutsConfig
	UseCasesTimeouts UseCasesTimeoutsConfig
}

type HandlerTimeoutsConfig struct {
	UserRegister time.Duration
	UserLogin    time.Duration
}

type UseCasesTimeoutsConfig struct { // Not implemented yet, but can be used for future use cases
	UserRegister time.Duration
}

func LoadConfig() (*ServerConfig, error) {
	resolver := path.NewResolver()

	cfg := &ServerConfig{
		Host:         helpers.Env("SERVER_HOST", "localhost"),
		Port:         helpers.Env("SERVER_PORT", "8080"),
		Environment:  helpers.Env("SERVER_ENVIRONMENT", "development"),
		APIContext:   helpers.Env("API_CONTEXT", "/api/v1"),
		TLSCertFile:  helpers.Env("SERVER_TLS_CERT_FILE", ""),
		TLSKeyFile:   helpers.Env("SERVER_TLS_KEY_FILE", ""),
		ReadTimeout:  helpers.GetEnvDuration("SERVER_READ_TIMEOUT", readTimeout),
		WriteTimeout: helpers.GetEnvDuration("SERVER_WRITE_TIMEOUT", writeTimeout),
		IdleTimeout:  helpers.GetEnvDuration("SERVER_IDLE_TIMEOUT", idleTimeout),
		Database: DatabaseConfig{
			Driver:              helpers.Env("DB_DRIVER", "sqlite3"),
			PostgresURL:         helpers.Env("PG_URL", "postgres://forum:password@localhost:5432/forumdb?sslmode=disable"),
			Path:                resolver.GetPath(helpers.Env("DB_PATH", "db/sqlite/data/forum.db")),
			MigrateOnStart:      helpers.GetEnvBool("DB_MIGRATE_ON_START", true),
			SeedOnStart:         helpers.GetEnvBool("DB_SEED_ON_START", true),
			Pragma_Foreign_Keys: helpers.Env("DB_PRAGMA_FOREIGN_KEYS", "_foreign_keys=on"),
			Pragma_Journal_Mode: helpers.Env("DB_PRAGMA_JOURNAL_MODE", "_journal_mode=WAL"),
			// Pragma:              fmt.Sprintf("_foreign_keys=%s&_journal_mode=%s",),
			OpenConn: helpers.GetEnvInt("DB_OPEN_CONN", 1),
		},
		SessionManager: SessionManagerConfig{
			DefaultExpiry:      helpers.GetEnvDuration("SESSION_DEFAULT_EXPIRY", defaultExpiry),
			SecureCookie:       helpers.GetEnvBool("SESSION_SECURE_COOKIE", false),
			AccessCookieName:   helpers.Env("SESSION_ACCESS_COOKIE_NAME", "access_token"),
			RefreshCookieName:  helpers.Env("SESSION_REFRESH_COOKIE_NAME", "refresh_token"),
			CookiePath:         helpers.Env("SESSION_COOKIE_PATH", "/"),
			CookieDomain:       helpers.Env("SESSION_COOKIE_DOMAIN", ""),
			HTTPOnlyCookie:     helpers.GetEnvBool("SESSION_HTTPONLY_COOKIE", true),
			SameSite:           helpers.Env("SESSION_SAMESITE", "Lax"),
			CleanupInterval:    helpers.GetEnvDuration("SESSION_CLEANUP_INTERVAL", cleanupInternal),
			MaxSessionsPerUser: helpers.GetEnvInt("SESSION_MAX_SESSIONS_PER_USER", maxSessionsPerUser),
			SessionIDLength:    helpers.GetEnvInt("SESSION_ID_LENGTH", sessionIDLenght),
			EnablePersistence:  helpers.GetEnvBool("SESSION_ENABLE_PERSISTENCE", true),
			LogSessions:        helpers.GetEnvBool("SESSION_LOG_SESSIONS", false),
			RefreshTokenExpiry: helpers.GetEnvDuration("SESSION_REFRESH_TOKEN_EXPIRY", refreshTokenExpiry),
		},
		Timeouts: TimeoutsConfig{
			HandlerTimeouts: HandlerTimeoutsConfig{
				UserRegister: helpers.GetEnvDuration("HANDLER_TIMEOUT_REGISTER", userRegisterTimeout),
				UserLogin:    helpers.GetEnvDuration("HANDLER_TIMEOUT_LOGIN", userLoginTimeout),
			},
		},
		OAuth: OAuthConfig{
			GitHub: GitHubOAuthConfig{
				ClientID:            helpers.Env("GITHUB_CLIENT_ID", ""),
				ClientSecret:        helpers.Env("GITHUB_CLIENT_SECRET", ""),
				RedirectURL:         helpers.Env("GITHUB_REDIRECT_URL", "http://localhost:8080/api/v1/auth/github/callback"),
				Scopes:              helpers.ParseList(helpers.Env("GITHUB_SCOPES", "user:email")),
				FrontendCallbackURL: helpers.Env("FRONTEND_GITHUB_CALLBACK_URL", "http://localhost:3001/auth/github/callback"),
			},
			Google: GoogleOAuthConfig{
				ClientID:            helpers.Env("GOOGLE_CLIENT_ID", ""),
				ClientSecret:        helpers.Env("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:         helpers.Env("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
				Scopes:              helpers.ParseList(helpers.Env("GOOGLE_SCOPES", "")),
				FrontendCallbackURL: helpers.Env("FRONTEND_GOOGLE_CALLBACK_URL", "http://localhost:3001/auth/google/callback"),
				TokenURL:            helpers.Env("GOOGLE_TOKEN_URL", ""),
			},
			FrontendCallbackURL: helpers.Env("FRONTEND_CALLBACK_URL", ""),
		},
		RateLimit: RateLimitConfig{
			Enabled:       helpers.GetEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsLimit: helpers.GetEnvInt("RATE_LIMIT_REQUESTS", defaultRateLimitRequestCapacity),
			WindowSeconds: int64(helpers.GetEnvInt("RATE_LIMIT_WINDOW_SECONDS", defaultRateLimitWindowSeconds)),
			Cleanup:       helpers.GetEnvDuration("RATE_LIMIT_CLEANUP_SECONDS", defaultRateLimitCleanupSeconds),
		},
	}

	if cfg.Host == "" {
		return nil, ErrMissingServerHost
	}
	_, err := strconv.Atoi(strings.TrimPrefix(cfg.Port, ":"))
	if err != nil {
		return nil, ErrServerPortNotInteger
	}

	return cfg, nil
}
