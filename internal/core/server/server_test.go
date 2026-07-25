package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"social-network/internal/config"
	"social-network/internal/core/server/servercontract"
)

// coreHandlerAdapter adapts *Server to servercontract.HandlerContract.
type coreHandlerAdapter struct {
	inner *Server
}

func (a *coreHandlerAdapter) Handler() http.Handler {
	return a.inner.Handler()
}

func (a *coreHandlerAdapter) Router() *http.ServeMux {
	return a.inner.Router()
}

func newCoreHandlerAdapter(t *testing.T) *coreHandlerAdapter {
	t.Helper()

	cfg := testConfig()
	srv := New(
		cfg,
		WithCORS(cfg.AllowedOrigins),
	)

	return &coreHandlerAdapter{inner: srv}
}

// coreLifecycleAdapter adapts *Server to servercontract.LifecycleContract.
// Uses a bare http.Server with the server's handler chain to test lifecycle
// without the ListenAndServe wrapper (signal handling tested separately).
type coreLifecycleAdapter struct {
	srv *http.Server
}

func (a *coreLifecycleAdapter) Start(t *testing.T) string {
	t.Helper()

	handler := newCoreHandlerAdapter(t).inner.Handler()

	a.srv = &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	lc := net.ListenConfig{}
	ctx := context.Background()
	ln, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	go func() {
		_ = a.srv.Serve(ln)
	}()

	time.Sleep(50 * time.Millisecond)
	return ln.Addr().String()
}

func (a *coreLifecycleAdapter) Shutdown(ctx context.Context) error {
	return a.srv.Shutdown(ctx)
}

// --- config helper ---

func testConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Host:           "127.0.0.1",
		Port:           "0",
		Environment:    "test",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		AllowedOrigins: []string{"*"},
		RateLimit: config.RateLimitConfig{
			Enabled:       false,
			RequestsLimit: 100,
			WindowSeconds: 60,
			Cleanup:       60 * time.Second,
		},
	}
}

// --- Handler contract tests ---

func TestContract_CORSHeaders(t *testing.T) {
	servercontract.TestCORSHeaders(t, newCoreHandlerAdapter(t))
}

func TestContract_CORSBlocksDisallowedOrigin(t *testing.T) {
	servercontract.TestCORSBlocksDisallowedOrigin(t, newCoreHandlerAdapter(t))
}

func TestContract_CORSHandlesPreflight(t *testing.T) {
	servercontract.TestCORSHandlesPreflight(t, newCoreHandlerAdapter(t))
}

func TestContract_RouteRegistration(t *testing.T) {
	servercontract.TestRouteRegistration(t, newCoreHandlerAdapter(t))
}

func TestContract_RouteRegistrationViaMux(t *testing.T) {
	servercontract.TestRouteRegistrationViaMux(t, newCoreHandlerAdapter(t))
}

func TestContract_SPAHandlerCatchAll(t *testing.T) {
	servercontract.TestSPAHandlerCatchAll(t, newCoreHandlerAdapter(t))
}

// --- Lifecycle contract tests ---

func TestContract_GracefulShutdown(t *testing.T) {
	servercontract.TestGracefulShutdown(t, &coreLifecycleAdapter{})
}

// --- Dedicated signal handling test ---

func TestServer_ListenAndServe_SignalShutdown(t *testing.T) {
	cfg := testConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := New(
		cfg,
		WithCORS(cfg.AllowedOrigins),
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected error on shutdown: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within 5s")
	}
}
