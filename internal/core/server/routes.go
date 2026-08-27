package server

import (
	"net/http"
)

const api = "/api/v1"

func RegisterRoutes(s *Server) {
	s.mux.HandleFunc(api+"/health", healthHandler)

	s.mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	if s.handlers == nil {
		return
	}

	require := s.requireAuth
	optional := s.optionalAuth

	// Realtime WebSocket
	if s.realtime != nil {
		s.mux.HandleFunc(api+"/ws", require(s.wsHandler))
	}

	// User routes
	if h := s.handlers.User; h != nil {
		s.mux.HandleFunc(api+"/register", h.Register)
		s.mux.HandleFunc(api+"/login/email", h.Login)
		s.mux.HandleFunc(api+"/login/username", h.Login)
		s.mux.HandleFunc(api+"/logout", require(h.Logout))
		s.mux.HandleFunc(api+"/me", require(h.GetMe))
		s.mux.HandleFunc(api+"/user/profile", optional(h.GetProfile))
		s.mux.HandleFunc(api+"/user/update", require(h.UpdateProfile))
		s.mux.HandleFunc(api+"/user/privacy", require(h.TogglePrivacy))
		s.mux.HandleFunc(api+"/user/activity", h.GetActivity)
		s.mux.HandleFunc(api+"/users", h.ListUsers)
	}

	// Follow routes
	if h := s.handlers.Follow; h != nil {
		s.mux.HandleFunc(api+"/follow", require(h.FollowUser))
		s.mux.HandleFunc(api+"/follow/unfollow", require(h.UnfollowUser))
		s.mux.HandleFunc(api+"/follow/accept", require(h.AcceptRequest))
		s.mux.HandleFunc(api+"/follow/decline", require(h.DeclineRequest))
		s.mux.HandleFunc(api+"/follow/followers", require(h.GetFollowers))
		s.mux.HandleFunc(api+"/follow/following", require(h.GetFollowing))
		s.mux.HandleFunc(api+"/follow/requests", require(h.GetPendingRequests))
		s.mux.HandleFunc(api+"/follow/connected", require(h.AreConnected))
	}

	// Chat routes
	if h := s.handlers.Chat; h != nil {
		s.mux.HandleFunc(api+"/chat/users", require(h.GetConversations))
		s.mux.HandleFunc(api+"/chat/history", require(h.GetChatHistory))
		s.mux.HandleFunc(api+"/chat/start", require(h.StartChat))
	}

	// Comment routes
	if h := s.handlers.Comment; h != nil {
		s.mux.HandleFunc(api+"/comments/create", require(h.CreateComment))
		s.mux.HandleFunc(api+"/comments/update", require(h.UpdateComment))
		s.mux.HandleFunc(api+"/comments/delete", require(h.DeleteComment))
		s.mux.HandleFunc(api+"/comments/vote", require(h.CastCommentVote))
		s.mux.HandleFunc(api+"/comments/vote/delete", require(h.DeleteCommentVote))
		s.mux.HandleFunc(api+"/comments/get", h.GetCommentByID)
		s.mux.HandleFunc(api+"/comments/topic", h.GetCommentsByTopic)
		s.mux.HandleFunc(api+"/comments/topic/votes", require(h.GetCommentsByTopicWithVotes))
		s.mux.HandleFunc(api+"/comments/votes/counts", require(h.GetVoteCounts))
	}

	// Topic routes
	if h := s.handlers.Topic; h != nil {
		s.mux.HandleFunc(api+"/topics/create", require(h.CreateTopic))
		s.mux.HandleFunc(api+"/topics/update", require(h.UpdateTopic))
		s.mux.HandleFunc(api+"/topics/delete", require(h.DeleteTopic))
		s.mux.HandleFunc(api+"/topics/vote", require(h.CastVote))
		s.mux.HandleFunc(api+"/topics/feed", optional(h.GetFeed))
		s.mux.HandleFunc(api+"/topics/get", optional(h.GetTopic))
		s.mux.HandleFunc(api+"/topics/user", optional(h.GetUserTopics))
		s.mux.HandleFunc(api+"/topics/group", optional(h.GetGroupTopics))
		s.mux.HandleFunc(api+"/topics/votes/counts", require(h.GetVoteCounts))
	}

	// OAuth routes
	if h := s.handlers.OAuth; h != nil {
		h.RegisterRoutes(s.mux)
	}

	// Group routes
	if h := s.handlers.Group; h != nil {
		s.mux.HandleFunc("POST "+api+"/groups", require(h.CreateGroup))
		s.mux.HandleFunc("GET "+api+"/groups", optional(h.ListGroups))
		s.mux.HandleFunc("GET "+api+"/groups/{groupId}", require(h.GetGroup))
		s.mux.HandleFunc("PUT "+api+"/groups/{groupId}", require(h.UpdateGroup))
		s.mux.HandleFunc("DELETE "+api+"/groups/{groupId}", require(h.DeleteGroup))
		s.mux.HandleFunc("DELETE "+api+"/groups/{groupId}/leave", require(h.LeaveGroup))
		s.mux.HandleFunc("POST "+api+"/groups/{groupId}/invite", require(h.InviteMember))
		s.mux.HandleFunc("POST "+api+"/groups/{groupId}/invite/respond", require(h.RespondInvite))
		s.mux.HandleFunc("GET "+api+"/groups/invitations/pending", require(h.GetPendingInvitations))
		s.mux.HandleFunc("GET "+api+"/groups/mine", require(h.ListMyGroups))
		s.mux.HandleFunc("GET "+api+"/groups/{groupId}/presence", require(h.GetGroupPresence))
		s.mux.HandleFunc("POST "+api+"/groups/{groupId}/request", require(h.RequestJoin))
		s.mux.HandleFunc("PUT "+api+"/groups/requests/{requestId}", require(h.RespondJoin))
		s.mux.HandleFunc("GET "+api+"/groups/{groupId}/posts", require(h.GetGroupFeed))
		s.mux.HandleFunc("POST "+api+"/groups/{groupId}/posts", require(h.CreateGroupPost))
		s.mux.HandleFunc("POST "+api+"/groups/posts/{postId}/vote", require(h.VoteGroupPost))
		s.mux.HandleFunc("GET "+api+"/groups/posts/{postId}/comments", require(h.GetGroupPostComments))
		s.mux.HandleFunc("POST "+api+"/groups/posts/{postId}/comments", require(h.CreateGroupPostComment))
		s.mux.HandleFunc("GET "+api+"/groups/{groupId}/members", require(h.GetGroupMembers))
		s.mux.HandleFunc("GET "+api+"/groups/{groupId}/invitations/sent", require(h.GetSentInvitationIDs))
		s.mux.HandleFunc("GET "+api+"/groups/{groupId}/requests/pending", require(h.GetPendingJoinRequests))
		s.mux.HandleFunc("GET "+api+"/groups/{groupId}/chat/messages", require(h.GetGroupChat))
	}

	// Event routes
	if h := s.handlers.Event; h != nil {
		s.mux.HandleFunc("POST "+api+"/groups/{groupId}/events", require(h.CreateEvent))
		s.mux.HandleFunc("GET "+api+"/groups/{groupId}/events", require(h.ListGroupEvents))
		s.mux.HandleFunc("PUT "+api+"/groups/{groupId}/events/{eventId}", require(h.UpdateEvent))
		s.mux.HandleFunc("POST "+api+"/events/{eventId}/respond", require(h.RespondToEvent))
		s.mux.HandleFunc("GET "+api+"/events/{eventId}/rsvps", require(h.ListEventResponders))
	}
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	if s.auth == nil {
		return next
	}
	return s.auth.Required(next)
}

func (s *Server) optionalAuth(next http.HandlerFunc) http.HandlerFunc {
	if s.auth == nil {
		return next
	}
	return s.auth.Optional(next)
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
