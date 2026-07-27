package server

func (s *Server) registerRoutes() {
	require := s.auth.Required

	s.mux.HandleFunc("GET /api/v1/notifications", require(s.notif.GetNotifications))
	s.mux.HandleFunc("GET /api/v1/notifications/unread-count", require(s.notif.GetUnreadCount))
	s.mux.HandleFunc("PATCH /api/v1/notifications/read", require(s.notif.MarkAsRead))
	s.mux.HandleFunc("PATCH /api/v1/notifications/read-all", require(s.notif.MarkAllAsRead))
	s.mux.HandleFunc("GET /api/v1/notifications/stream", require(s.notif.StreamNotifications))
}
