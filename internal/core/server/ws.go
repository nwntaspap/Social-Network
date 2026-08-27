package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"social-network/internal/core/middleware"
	"social-network/internal/core/realtime"
)

type realtimeServer struct {
	hub            *realtime.Hub
	router         realtime.WSRouter
	allowedOrigins map[string]bool
	wildcard       bool
}

// WithRealtime registers the WebSocket hub and router with the server.
// The allowed origins are validated against the configured CORS origins.
func WithRealtime(hub *realtime.Hub, router realtime.WSRouter, allowedOrigins []string) Option {
	origins := make(map[string]bool, len(allowedOrigins))
	wildcard := false
	for _, o := range allowedOrigins {
		if o == "*" {
			wildcard = true
			continue
		}
		origins[o] = true
	}
	return func(s *Server) {
		s.realtime = &realtimeServer{hub: hub, router: router, allowedOrigins: origins, wildcard: wildcard}
	}
}

func (r *realtimeServer) checkOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	return r.wildcard || r.allowedOrigins[origin]
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	if s.realtime == nil {
		http.NotFound(w, r)
		return
	}

	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(req *http.Request) bool {
			return s.realtime.checkOrigin(req.Header.Get("Origin"))
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade error: %v", err)
		return
	}

	client := realtime.NewClient(userID, s.realtime.hub, conn)
	s.realtime.hub.Register(client)

	go client.WritePump()
	go client.ReadPump(func(c *realtime.Client, msg []byte) {
		s.realtime.router.Route(c, msg)
	})
}
