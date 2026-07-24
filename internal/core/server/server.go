package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"social-network/internal/config"
	"social-network/internal/core/middleware"
	"social-network/internal/core/middleware/ratelimiter"
	"social-network/internal/core/session"

	chattransport "social-network/internal/chat/transport"
	commenttransport "social-network/internal/comment/transport"
	followtransport "social-network/internal/follow/transport"
	oauthtransport "social-network/internal/oauth/transport"
	topictransport "social-network/internal/topic/transport"
	usertransport "social-network/internal/user/transport"
)

type AllHandlers struct {
	User    *usertransport.Handler
	Follow  *followtransport.Handler
	Chat    *chattransport.Handler
	Comment *commenttransport.Handler
	Topic   *topictransport.Handler
	OAuth   *oauthtransport.Handler
}

type Server struct {
	mux    *http.ServeMux
	srv    *http.Server
	cfg    *config.ServerConfig
	auth   *middleware.Auth
	logger *slog.Logger

	allowedOrigins []string
	rateLimiterOpt *rateLimiterOptions

	handlers *AllHandlers

	handler http.Handler
}

type rateLimiterOptions struct {
	window *ratelimiter.Window
	limit  int
}

type Option func(*Server)

func WithAuth(sm session.Manager, cookieName string) Option {
	return func(s *Server) {
		s.auth = middleware.NewAuth(sm, cookieName)
	}
}

func WithCORS(allowedOrigins []string) Option {
	return func(s *Server) {
		s.allowedOrigins = allowedOrigins
	}
}

func WithLogging(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

func WithRateLimiter(window *ratelimiter.Window, limit int) Option {
	return func(s *Server) {
		s.rateLimiterOpt = &rateLimiterOptions{window: window, limit: limit}
	}
}

func WithHandlers(h *AllHandlers) Option {
	return func(s *Server) {
		s.handlers = h
	}
}

func New(cfg *config.ServerConfig, opts ...Option) *Server {
	s := &Server{
		mux: http.NewServeMux(),
		cfg: cfg,
	}
	for _, opt := range opts {
		opt(s)
	}
	RegisterRoutes(s)
	s.buildHandler()
	return s
}

func (s *Server) buildHandler() {
	var h http.Handler = s.mux

	if s.rateLimiterOpt != nil {
		h = middleware.NewRateLimiter(s.rateLimiterOpt.window, s.rateLimiterOpt.limit, h)
	}
	if s.allowedOrigins != nil {
		h = middleware.NewCORS(s.allowedOrigins, h)
	}
	if s.logger != nil {
		h = middleware.NewLogging(s.logger, h)
	}

	s.handler = h
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) Router() *http.ServeMux {
	return s.mux
}

func (s *Server) Auth() *middleware.Auth {
	return s.auth
}

func (s *Server) logShutdown(msg string) {
	if s.logger != nil {
		s.logger.Info(msg)
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	addr := s.cfg.Host + ":" + s.cfg.Port

	s.srv = &http.Server{
		Addr:              addr,
		Handler:           s.handler,
		ReadHeaderTimeout: s.cfg.ReadTimeout,
		ReadTimeout:       s.cfg.ReadTimeout,
		WriteTimeout:      s.cfg.WriteTimeout,
		IdleTimeout:       s.cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
		if s.cfg.TLSCertFile != "" && s.cfg.TLSKeyFile != "" {
			log.Println("HTTPS Server Listening on port:", s.cfg.Port)
			errCh <- s.srv.ListenAndServeTLS(s.cfg.TLSCertFile, s.cfg.TLSKeyFile)
		} else {
			log.Println("HTTP Server Listening on port:", s.cfg.Port)
			errCh <- s.srv.ListenAndServe()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-sigCh:
		s.logShutdown("received SIGINT/SIGTERM, shutting down")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}

	signal.Stop(sigCh)

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	s.logShutdown("shutting down gracefully")

	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	s.logShutdown("server stopped")

	return nil
}
