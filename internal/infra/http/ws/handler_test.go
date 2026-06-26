package ws

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"social-network/internal/domain/user"
	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	wsPkg "social-network/internal/infra/ws"

	"github.com/gorilla/websocket"
)

type stubRouter struct{}

func (s *stubRouter) Route(_ *wsPkg.Client, _ []byte) {}

func testHandler(t *testing.T, allowed []string) *Handler {
	t.Helper()
	return NewHandler(wsPkg.NewHub(), &stubRouter{}, logger.New(io.Discard, logger.LevelOff), allowed)
}

func testServer(t *testing.T, h *Handler) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.Key("user"), &user.User{ID: "u1"})
		h.UpgradeConnection(w, r.WithContext(ctx))
	}))
}

func TestHandler_CheckOrigin_AllowsConfiguredOrigins(t *testing.T) {
	h := testHandler(t, []string{"https://trusted.com"})
	server := testServer(t, h)
	defer server.Close()

	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	_, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Origin": []string{"https://trusted.com"},
	})
	if err != nil {
		t.Fatalf("allowed origin should connect, got error: %v", err)
	}
}

func TestHandler_CheckOrigin_RejectsUnknownOrigin(t *testing.T) {
	h := testHandler(t, []string{"https://trusted.com"})
	server := testServer(t, h)
	defer server.Close()

	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	_, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Origin": []string{"https://evil.com"},
	})
	if err == nil {
		t.Fatal("unknown origin should be rejected, but dial succeeded")
	}
}

func TestHandler_NewHandler_StoresAllowedOrigins(t *testing.T) {
	h := testHandler(t, []string{"https://a.com", "https://b.com"})

	if h.allowedOrigins == nil {
		t.Fatal("allowedOrigins should not be nil")
	}
	if len(h.allowedOrigins) != 2 {
		t.Fatalf("allowedOrigins len = %d, want 2", len(h.allowedOrigins))
	}
	if h.allowedOrigins[0] != "https://a.com" {
		t.Errorf("allowedOrigins[0] = %q, want %q", h.allowedOrigins[0], "https://a.com")
	}
}
