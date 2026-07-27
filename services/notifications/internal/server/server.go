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

	"social-network/services/notifications/internal/config"
	"social-network/services/notifications/internal/middleware"
)

type NotificationsHandler interface {
	GetNotifications(w http.ResponseWriter, r *http.Request)
	GetUnreadCount(w http.ResponseWriter, r *http.Request)
	MarkAsRead(w http.ResponseWriter, r *http.Request)
	MarkAllAsRead(w http.ResponseWriter, r *http.Request)
	StreamNotifications(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	mux    *http.ServeMux
	srv    *http.Server
	cfg    *config.Config
	auth   *middleware.Auth
	logger *slog.Logger
	notif  NotificationsHandler

	handler http.Handler
}

func New(cfg *config.Config, logger *slog.Logger, auth *middleware.Auth, notif NotificationsHandler) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		cfg:    cfg,
		logger: logger,
		auth:   auth,
		notif:  notif,
	}
	s.registerRoutes()
	s.buildHandler()
	return s
}

func (s *Server) buildHandler() {
	var h http.Handler = s.mux

	h = middleware.NewRateLimiter(100, time.Minute, h)

	if s.logger != nil {
		h = middleware.NewLogging(s.logger, h)
	}

	s.handler = h
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
		log.Println("Notifications service listening on", addr)
		errCh <- s.srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-sigCh:
		log.Println("received SIGINT/SIGTERM, shutting down")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}

	signal.Stop(sigCh)

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	log.Println("shutting down gracefully")
	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	log.Println("server stopped")

	return nil
}
