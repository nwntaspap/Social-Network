package bootstrap

import (
	"context"
	"net/http"

	commentcommands "social-network/internal/comment/commands"
	commentqueries "social-network/internal/comment/queries"
	commentstore "social-network/internal/comment/store"
	commenttransport "social-network/internal/comment/transport"
	"social-network/internal/core/middleware"
	"social-network/internal/platform/database"
)

type commentEventBus struct{}

func (b *commentEventBus) Publish(_ context.Context, _ string, _ any) error {
	return nil
}

func initComment(db database.DB) *commenttransport.Handler {
	store := commentstore.NewSQLiteStore(db)
	bus := &commentEventBus{}

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	return commenttransport.NewHandler(
		extractUser,
		commentcommands.NewCreateCommentHandler(store, bus),
		commentcommands.NewUpdateCommentHandler(store),
		commentcommands.NewDeleteCommentHandler(store),
		commentcommands.NewCastCommentVoteHandler(store, bus),
		commentqueries.NewGetCommentByIDResolver(store),
		commentqueries.NewGetCommentByIDWithVotesResolver(store),
		commentqueries.NewGetCommentsByTopicResolver(store),
		commentqueries.NewGetCommentsByTopicWithVotesResolver(store),
		commentqueries.NewGetVoteCountsResolver(store),
	)
}
