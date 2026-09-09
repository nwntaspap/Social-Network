package bootstrap

import (
	"maps"

	"social-network/internal/chat"
	chatcommands "social-network/internal/chat/commands"
	"social-network/internal/chat/queries"
	chatstore "social-network/internal/chat/store"
	chattransport "social-network/internal/chat/transport"
	"social-network/internal/core/realtime"
	followstore "social-network/internal/follow/store"
	groupcommands "social-network/internal/group/commands"
	groupqueries "social-network/internal/group/queries"
	groupstore "social-network/internal/group/store"
	grouptransport "social-network/internal/group/transport"
	"social-network/internal/platform/database"
	"social-network/internal/platform/logger"
)

// Realtime bundles the WebSocket hub and its message router.
type Realtime struct {
	Hub    *realtime.Hub
	Router realtime.WSRouter
}

func initRealtime(db database.DB, hub *realtime.Hub, logger logger.Logger) *Realtime {
	chatStore := chatstore.NewSQLiteStore(db)
	followStore := followstore.NewSQLiteStore(db)

	followAdapter := &chat.FollowAdapter{AreConnectedFn: followStore.AreConnected}

	gate := chatcommands.NewMessageGate(followAdapter)
	send := chatcommands.NewSendPrivateMessageHandler(chatStore, gate)
	getHistory := queries.NewGetChatHistoryResolver(chatStore, followAdapter)

	chatWS := chattransport.NewWSHandler(hub, send, getHistory, chatStore, chatStore, gate)

	groupStore := groupstore.NewSQLiteStore(db)
	groupWS := grouptransport.NewGroupWSHandler(
		hub,
		groupcommands.NewSendGroupMessageHandler(groupStore),
		groupqueries.NewGetGroupChatResolver(groupStore),
		groupqueries.NewListGroupMemberIDsResolver(groupStore),
		groupcommands.NewMarkGroupReadHandler(groupStore),
		logger,
	)

	handlers := chatWS.Handlers()
	maps.Copy(handlers, groupWS.Handlers())

	return &Realtime{
		Hub:    hub,
		Router: realtime.NewWSRouter(handlers),
	}
}
