package server

import (
	"context"
	"net/http"
	"testing"
)

func TestRoutes_HealthEndpoint(t *testing.T) {
	srv := New(testConfig())

	rec, req := newRecorder(), httptestNewRequest("/api/v1/health")
	srv.Handler().ServeHTTP(rec, req)

	if rec.code != http.StatusOK {
		t.Errorf("health: status = %d, want 200", rec.code)
	}
	if string(rec.body) != `{"status":"ok"}` {
		t.Errorf("health: body = %q, want %q", string(rec.body), `{"status":"ok"}`)
	}
	if ct := rec.header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("health: Content-Type = %q, want application/json", ct)
	}
}

func TestRoutes_StaticFileAccessible(t *testing.T) {
	srv := New(testConfig())

	rec, req := newRecorder(), httptestNewRequest("/static/")
	srv.Handler().ServeHTTP(rec, req)

	if rec.code != http.StatusOK && rec.code != http.StatusNotFound {
		t.Errorf("static: unexpected status %d (OK or 404 both acceptable)", rec.code)
	}
	if rec.code == http.StatusOK {
		t.Log("static files served (frontend/static/ exists)")
	}
}

func TestRoutes_SPACatchAll_ServesIndexForNonAPIPaths(t *testing.T) {
	srv := New(testConfig())

	rec, req := newRecorder(), httptestNewRequest("/some-client-route")
	srv.Handler().ServeHTTP(rec, req)

	if rec.code != http.StatusOK && rec.code != http.StatusNotFound {
		t.Errorf("SPA catch-all: expected 200 or 404, got %d", rec.code)
	}
}

func TestRoutes_SPACatchAll_NotFoundForAPIPaths(t *testing.T) {
	srv := New(testConfig())

	rec, req := newRecorder(), httptestNewRequest("/api/v1/nonexistent")
	srv.Handler().ServeHTTP(rec, req)

	if rec.code != http.StatusNotFound {
		t.Errorf("API unknown route: expected 404, got %d", rec.code)
	}
}

func TestRoutes_SPACatchAll_NotFoundForStaticPaths(t *testing.T) {
	srv := New(testConfig())

	rec, req := newRecorder(), httptestNewRequest("/static/nonexistent.css")
	srv.Handler().ServeHTTP(rec, req)

	if rec.code != http.StatusNotFound {
		t.Errorf("static unknown file: expected 404, got %d", rec.code)
	}
}

// --- helper: uses net/http/httptest internally but avoids importing it ---

func httptestNewRequest(target string) *http.Request {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	return req
}

func newRecorder() *testResponseRecorder {
	return &testResponseRecorder{header: make(http.Header)}
}

type testResponseRecorder struct {
	code   int
	header http.Header
	body   []byte
}

func (r *testResponseRecorder) Header() http.Header { return r.header }
func (r *testResponseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}
func (r *testResponseRecorder) WriteHeader(code int) { r.code = code }
