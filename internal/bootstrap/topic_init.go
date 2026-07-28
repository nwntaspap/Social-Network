package bootstrap

import (
	"context"
	"net/http"

	"social-network/internal/core/middleware"
	localstorage "social-network/internal/infra/storage/local"
	"social-network/internal/platform/database"
	topiccommands "social-network/internal/topic/commands"
	topicqueries "social-network/internal/topic/queries"
	topicstore "social-network/internal/topic/store"
	topictransport "social-network/internal/topic/transport"
)

func initTopic(db database.DB) *topictransport.Handler {
	store := topicstore.NewSQLiteStore(db)
	bus := &topicEventBus{}
	img := localstorage.NewLocalStorage()

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	return topictransport.NewHandler(
		extractUser,
		topiccommands.NewCreateTopicHandler(store, bus, img),
		topiccommands.NewUpdateTopicHandler(store, img),
		topiccommands.NewDeleteTopicHandler(store, bus, img),
		topiccommands.NewCastVoteHandler(store, bus),
		topicqueries.NewGetFeedResolver(store),
		topicqueries.NewGetTopicResolver(store),
		topicqueries.NewGetTopicsByUserResolver(store),
		topicqueries.NewGetTopicsByGroupResolver(store),
		topicqueries.NewGetVoteCountsResolver(store),
	)
}

type topicEventBus struct{}

func (b *topicEventBus) Publish(_ context.Context, _ string, _ any) error {
	return nil
}
