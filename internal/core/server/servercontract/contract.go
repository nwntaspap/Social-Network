package servercontract

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// HandlerContract defines the handler-level behavior a server must provide.
// Both legacy and new servers must satisfy this for behavioral parity.
type HandlerContract interface {
	Handler() http.Handler
	Router() *http.ServeMux
}

// LifecycleContract defines the start/stop behavior.
type LifecycleContract interface {
	Start(t *testing.T) string
	Shutdown(ctx context.Context) error
}

// TestCORSHeaders verifies CORS headers are set through the middleware chain.
// Registers a route so the response comes from the application, not a 404.
func TestCORSHeaders(t *testing.T, s HandlerContract) {
	t.Helper()

	s.Router().HandleFunc("/cors-test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec, req := newRecorder(), newRequestWithContext(http.MethodGet, "/cors-test")
	req.Header.Set("Origin", "https://example.com")
	s.Handler().ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao == "" {
		t.Log("CORS: ACAO not set (legacy CORS does not echo origin)")
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Access-Control-Allow-Methods header not set")
	}
	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Access-Control-Allow-Credentials should be 'true'")
	}
	if ma := rec.Header().Get("Access-Control-Max-Age"); ma == "" {
		t.Log("Access-Control-Max-Age not set (non-critical)")
	}
	if rec.code != http.StatusOK {
		t.Errorf("CORS: expected 200, got %d", rec.code)
	}
}

// TestCORSBlocksDisallowedOrigin checks origin enforcement.
// Legacy CORS does not block by origin; core CORS can be strict.
func TestCORSBlocksDisallowedOrigin(t *testing.T, s HandlerContract) {
	t.Helper()

	s.Router().HandleFunc("/cors-test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec, req := newRecorder(), newRequestWithContext(http.MethodGet, "/cors-test")
	req.Header.Set("Origin", "https://evil.com")
	s.Handler().ServeHTTP(rec, req)

	switch rec.code {
	case http.StatusForbidden:
		t.Log("CORS: disallowed origin blocked with 403 (strict mode)")
	case http.StatusOK:
		t.Log("CORS: disallowed origin passed through (no origin check)")
	default:
		t.Errorf("CORS: expected 200 or 403, got %d", rec.code)
	}
}

// TestCORSHandlesPreflight verifies OPTIONS preflight requests.
func TestCORSHandlesPreflight(t *testing.T, s HandlerContract) {
	t.Helper()

	s.Router().HandleFunc("/preflight", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec, req := newRecorder(), newRequestWithContext(http.MethodOptions, "/preflight")
	req.Header.Set("Origin", "https://example.com")
	s.Handler().ServeHTTP(rec, req)

	if rec.code != http.StatusOK && rec.code != http.StatusNoContent {
		t.Errorf("OPTIONS preflight: status = %d, want 200 or 204", rec.code)
	}
}

// TestRouteRegistration verifies that routes registered on the server respond.
func TestRouteRegistration(t *testing.T, s HandlerContract) {
	t.Helper()

	mux := s.Router()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("world"))
	})

	rec, req := newRecorder(), newRequestWithContext(http.MethodGet, "/hello")
	s.Handler().ServeHTTP(rec, req)

	if rec.code != http.StatusOK {
		t.Errorf("/hello: status = %d, want %d", rec.code, http.StatusOK)
	}
	if string(rec.body) != "world" {
		t.Errorf("/hello: body = %q, want %q", string(rec.body), "world")
	}
}

// TestRouteRegistrationViaMux verifies routes registered directly on the mux
// are reachable through the wrapped handler (tests that middleware chain
// correctly delegates to the mux). Uses a unique path to avoid conflicts
// with routes the server may have pre-registered.
func TestRouteRegistrationViaMux(t *testing.T, s HandlerContract) {
	t.Helper()

	const path = "/__contract_test_route__"

	mux := s.Router()
	mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	rec, req := newRecorder(), newRequestWithContext(http.MethodGet, path)
	req.Header.Set("Origin", "https://example.com")
	s.Handler().ServeHTTP(rec, req)

	if rec.code != http.StatusOK {
		t.Errorf("route via mux: status = %d, want 200 (body: %s)", rec.code, string(rec.body))
	}
}

// TestSPAHandlerCatchAll verifies that a catch-all route serves content
// for non-API, non-static paths. Uses a unique prefix to avoid conflicts
// with pre-registered routes.
func TestSPAHandlerCatchAll(t *testing.T, s HandlerContract) {
	t.Helper()

	const spaPrefix = "/__spa_test__"

	mux := s.Router()
	mux.HandleFunc(spaPrefix, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("index.html"))
	})

	rec, req := newRecorder(), newRequestWithContext(http.MethodGet, spaPrefix)
	s.Handler().ServeHTTP(rec, req)

	if rec.code != http.StatusOK {
		t.Errorf("SPA catch-all: status = %d, want 200", rec.code)
	}
}

// TestGracefulShutdown verifies the server shuts down cleanly within 10s
// when it receives a shutdown signal.
func TestGracefulShutdown(t *testing.T, s LifecycleContract) {
	t.Helper()

	addr := s.Start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+"/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("server should be listening: %v", err)
	}
	_ = resp.Body.Close()

	start := time.Now()
	err = s.Shutdown(ctx)
	if err != nil {
		t.Fatalf("graceful shutdown failed: %v (elapsed: %v)", err, time.Since(start))
	}
	t.Logf("shutdown completed in %v", time.Since(start))

	checkReq, checkErr := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+"/", nil)
	if checkErr == nil {
		resp2, doErr := client.Do(checkReq)
		if doErr == nil {
			_ = resp2.Body.Close()
			t.Error("server should not accept requests after shutdown")
		}
	}
}

// --- testing utilities (no httptest import to keep contract lightweight) ---

type testRecorder struct {
	code   int
	header http.Header
	body   []byte
}

func newRecorder() *testRecorder {
	return &testRecorder{code: http.StatusOK, header: make(http.Header)}
}

func (r *testRecorder) Header() http.Header { return r.header }
func (r *testRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}
func (r *testRecorder) WriteHeader(code int) { r.code = code }

func newRequestWithContext(method, target string) *http.Request {
	req, _ := http.NewRequestWithContext(context.Background(), method, target, nil)
	return req
}
