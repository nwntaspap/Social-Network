package bootstrap

import (
	"context"
	"net/http"

	commentstore "social-network/internal/comment/store"
	"social-network/internal/core/middleware"
	coresessionstore "social-network/internal/core/session/store"
	followstore "social-network/internal/follow/store"
	"social-network/internal/pkg/bcrypt"
	"social-network/internal/pkg/uuid"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
	"social-network/internal/platform/logger"
	localstorage "social-network/internal/platform/storage/local"
	topicstore "social-network/internal/topic/store"
	usercommands "social-network/internal/user/commands"
	userqueries "social-network/internal/user/queries"
	userstore "social-network/internal/user/store"
	usertransport "social-network/internal/user/transport"
)

func initUser(db database.DB, sessionStore *coresessionstore.Store, cookies *middleware.SessionCookies, isOnline func(string) bool, bus eventbus.EventBus, log logger.Logger) *usertransport.Handler {
	userStore := userstore.NewSQLiteStore(db)
	followStore := followstore.NewSQLiteStore(db)
	topicStore := topicstore.NewSQLiteStore(db)
	commentStore := commentstore.NewSQLiteStore(db)

	uuidProvider := uuid.NewProvider()
	bcryptProvider := bcrypt.NewProvider()

	fc := &followCheckerAdapter{store: followStore}

	return usertransport.NewHandler(
		&authUserExtractor{},
		usercommands.NewRegisterHandler(userStore, uuidProvider, bcryptProvider, localstorage.NewLocalStorage()),
		usercommands.NewLoginHandler(userStore, bcryptProvider, sessionStore),
		usercommands.NewLogoutHandler(sessionStore),
		usercommands.NewUpdateProfileHandler(userStore, bus),
		usercommands.NewTogglePrivacyHandler(userStore),
		userqueries.NewGetProfileResolver(userStore, fc, followStore),
		userqueries.NewGetActivityResolver(userStore, topicStore, commentStore, topicStore, followStore),
		userqueries.NewListUsersResolver(userStore, isOnline),
		cookies,
		log,
	)
}

type authUserExtractor struct{}

func (e *authUserExtractor) Extract(r *http.Request) (string, bool) {
	uid := middleware.GetUserIDFromContext(r)
	if uid == "" {
		return "", false
	}
	return uid, true
}

type followCheckerAdapter struct {
	store *followstore.SQLiteStore
}

func (a *followCheckerAdapter) IsFollowing(ctx context.Context, followerID, targetID string) (bool, error) {
	return a.store.AreConnected(ctx, followerID, targetID)
}

var _ userqueries.FollowChecker = (*followCheckerAdapter)(nil)
