package bootstrap

import (
	"database/sql"
	"net/http"

	"social-network/internal/follow"
	followcommands "social-network/internal/follow/commands"
	followqueries "social-network/internal/follow/queries"
	followstore "social-network/internal/follow/store"
	followtransport "social-network/internal/follow/transport"
	"social-network/internal/infra/middleware"
)

func initFollow(db *sql.DB) *followtransport.Handler {
	store := followstore.NewSQLiteStore(db)
	privacy := &follow.PrivacyStub{}
	bus := &follow.NoopEventBus{}

	extractUser := func(r *http.Request) (string, bool) {
		user := middleware.GetUserFromContext(r)
		if user == nil {
			return "", false
		}
		return user.ID, true
	}

	return followtransport.NewHandler(
		extractUser,
		followcommands.NewFollowUserHandler(store, privacy, bus),
		followcommands.NewUnfollowUserHandler(store),
		followcommands.NewAcceptRequestHandler(store, bus),
		followcommands.NewDeclineRequestHandler(store, bus),
		followqueries.NewGetFollowersResolver(store),
		followqueries.NewGetFollowingResolver(store),
		followqueries.NewGetPendingRequestsResolver(store),
		followqueries.NewAreConnectedResolver(store),
	)
}
