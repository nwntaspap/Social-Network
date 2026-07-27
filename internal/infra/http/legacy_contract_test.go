package http

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"social-network/internal/bootstrap"
	"social-network/internal/config"
	"social-network/internal/core/server/servercontract"
	domainsession "social-network/internal/domain/session"
	domainuser "social-network/internal/domain/user"
	"social-network/internal/infra/http/authcookies"
	"social-network/internal/infra/logger"
	inframiddleware "social-network/internal/infra/middleware"
	"social-network/internal/infra/realtime/notifications"
	"social-network/internal/infra/ws"
	oauth "social-network/internal/pkg/oAuth"
	"social-network/internal/pkg/oAuth/githubclient"
	"social-network/internal/pkg/oAuth/googleclient"
)

// --- mock session.Manager ---

type mockSessionManager struct{}

func (m *mockSessionManager) CreateSession(_ context.Context, _ string) (*domainsession.Session, error) {
	return nil, errors.New("mock: not available")
}

func (m *mockSessionManager) GetSession(_ string) (*domainsession.Session, error) {
	return nil, errors.New("mock: not available")
}

func (m *mockSessionManager) DeleteSession(_ string) error {
	return errors.New("mock: not available")
}

func (m *mockSessionManager) GetUserFromSession(_ string) (*domainuser.User, error) {
	return nil, errors.New("mock: not available")
}

func (m *mockSessionManager) GetSessionFromSessionTokens(_, _ string) (*domainsession.Session, error) {
	return nil, errors.New("mock: not available")
}

func (m *mockSessionManager) GetSessionByRefreshToken(_ string) (*domainsession.Session, error) {
	return nil, errors.New("mock: not available")
}

func (m *mockSessionManager) ValidateSession(_ string) error {
	return errors.New("mock: not available")
}

func (m *mockSessionManager) DeleteSessionWhenNewCreated(_ context.Context, _, _ string) error {
	return errors.New("mock: not available")
}

// --- legacy adapter for HandlerContract ---

type legacyHandlerAdapter struct {
	mux    *http.ServeMux
	config *config.ServerConfig
}

func (a *legacyHandlerAdapter) Handler() http.Handler {
	handler := http.Handler(a.mux)
	handler = inframiddleware.NewCorsMiddleware(handler)
	if a.config.RateLimit.Enabled {
		// rate limiter wrapping omitted for test simplicity
	}
	return handler
}

func (a *legacyHandlerAdapter) Router() *http.ServeMux {
	return a.mux
}

func newLegacyHandlerAdapter(t *testing.T) *legacyHandlerAdapter {
	t.Helper()

	cfg := testConfig()
	srv := newLegacyServer(t, cfg)

	return &legacyHandlerAdapter{
		mux:    srv.router,
		config: cfg,
	}
}

// --- legacy adapter for LifecycleContract ---

type legacyLifecycleAdapter struct {
	srv  *Server
	http *http.Server
}

func (a *legacyLifecycleAdapter) Start(t *testing.T) string {
	t.Helper()

	handler := inframiddleware.NewCorsMiddleware(a.srv.router)
	a.http = &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: handler,
	}

	ln, err := net.Listen("tcp", a.http.Addr)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	go func() {
		_ = a.http.Serve(ln)
	}()

	time.Sleep(50 * time.Millisecond)
	return ln.Addr().String()
}

func (a *legacyLifecycleAdapter) Shutdown(ctx context.Context) error {
	return a.http.Shutdown(ctx)
}

// --- helpers ---

func testConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Host:           "127.0.0.1",
		Port:           "0",
		Environment:    "test",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		AllowedOrigins: []string{"*"},
		RateLimit:      config.RateLimitConfig{Enabled: false},
		SessionManager: config.SessionManagerConfig{
			AccessCookieName:  "access_token",
			RefreshCookieName: "refresh_token",
			CookiePath:        "/",
			SameSite:          "Lax",
		},
	}
}

func newLegacyServer(t *testing.T, cfg *config.ServerConfig) *Server {
	t.Helper()

	sm := &mockSessionManager{}
	cm := authcookies.NewManager(cfg.SessionManager)
	mw := inframiddleware.NewMiddleware(sm, cm)
	log := logger.New(io.Discard, logger.LevelOff)
	hub := ws.NewHub()
	ntf := notifications.NewNotifier()

	minOAuth := &oauth.OAuth{
		StateManager:   oauth.NewStateManager(time.Minute),
		GithubProvider: githubclient.NewProvider("", "", "", nil),
		GoogleProvider: googleclient.NewProvider("", "", "", "", nil),
	}

	app := &bootstrap.App{
		Middlware:      mw,
		SessionManager: sm,
		CookieManager:  cm,
		Logger:         log,
		Hub:            hub,
		Notifier:       ntf,
		OAuth:          nil,
		LegacyOAuth:    minOAuth,
		// Follow, FileStorage, Services — zero-valued,
		// registered routes will panic if invoked but contract tests
		// register their own test routes on the mux.
	}

	srv := NewServer(cfg, app)
	return srv
}

// --- Handler contract tests ---

func TestLegacyContract_CORSHeaders(t *testing.T) {
	servercontract.TestCORSHeaders(t, newLegacyHandlerAdapter(t))
}

func TestLegacyContract_CORSBlocksDisallowedOrigin(t *testing.T) {
	servercontract.TestCORSBlocksDisallowedOrigin(t, newLegacyHandlerAdapter(t))
}

func TestLegacyContract_CORSHandlesPreflight(t *testing.T) {
	servercontract.TestCORSHandlesPreflight(t, newLegacyHandlerAdapter(t))
}

func TestLegacyContract_RouteRegistration(t *testing.T) {
	servercontract.TestRouteRegistration(t, newLegacyHandlerAdapter(t))
}

func TestLegacyContract_RouteRegistrationViaMux(t *testing.T) {
	servercontract.TestRouteRegistrationViaMux(t, newLegacyHandlerAdapter(t))
}

func TestLegacyContract_SPAHandlerCatchAll(t *testing.T) {
	servercontract.TestSPAHandlerCatchAll(t, newLegacyHandlerAdapter(t))
}

// --- Lifecycle contract tests (graceful shutdown is new-server only) ---

func TestLegacyContract_GracefulShutdown(t *testing.T) {
	cfg := testConfig()
	srv := newLegacyServer(t, cfg)
	adapter := &legacyLifecycleAdapter{srv: srv}
	servercontract.TestGracefulShutdown(t, adapter)
}
