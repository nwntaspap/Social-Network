package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_AllowedOrigin(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	cors := NewCORS([]string{"https://example.com"}, next)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	cors.ServeHTTP(w, r)

	if !called {
		t.Fatal("downstream handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("ACAO = %q, want %q", got, "https://example.com")
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("ACAC = %q, want %q", got, "true")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("downstream handler should not be called")
	})

	cors := NewCORS([]string{"https://example.com"}, next)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()

	cors.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cors := NewCORS([]string{"https://example.com"}, next)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	cors.ServeHTTP(w, r)

	if !called {
		t.Fatal("downstream handler was not called for same-origin request")
	}
}

func TestCORS_PreflightOptions(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cors := NewCORS([]string{"https://example.com"}, next)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	cors.ServeHTTP(w, r)

	if called {
		t.Fatal("downstream handler should not be called for OPTIONS preflight")
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("ACAM = %q, want full methods list", got)
	}
}

func TestCORS_WildcardOrigin(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cors := NewCORS([]string{"*"}, next)

	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://any-origin.com")
	w := httptest.NewRecorder()

	cors.ServeHTTP(w, r)

	if !called {
		t.Fatal("downstream handler was not called")
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://any-origin.com" {
		t.Errorf("ACAO = %q, want %q", got, "https://any-origin.com")
	}
}

func TestCORS_MultipleAllowedOrigins(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cors := NewCORS([]string{"https://a.com", "https://b.com"}, next)

	// Test first allowed origin
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://a.com")
	w := httptest.NewRecorder()
	cors.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("origin a: status = %d, want %d", w.Code, http.StatusOK)
	}

	// Test second allowed origin
	r = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://b.com")
	w = httptest.NewRecorder()
	cors.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("origin b: status = %d, want %d", w.Code, http.StatusOK)
	}

	// Test disallowed origin
	r = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://c.com")
	w = httptest.NewRecorder()
	cors.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("origin c: status = %d, want %d", w.Code, http.StatusForbidden)
	}
}
