package bootstrap

import (
	"context"
	"net/http"

	commentstore "social-network/internal/comment/store"
	"social-network/internal/core/middleware"
	coresession "social-network/internal/core/session"
	followstore "social-network/internal/follow/store"
	"social-network/internal/pkg/bcrypt"
	"social-network/internal/pkg/uuid"
	"social-network/internal/platform/database"
	topicstore "social-network/internal/topic/store"
	usercommands "social-network/internal/user/commands"
	userqueries "social-network/internal/user/queries"
	userstore "social-network/internal/user/store"
	usertransport "social-network/internal/user/transport"
)

func initUser(db database.DB, sessionMgr *coreSessionAdapter) *usertransport.Handler {
	userStore := userstore.NewSQLiteStore(db)
	followStore := followstore.NewSQLiteStore(db)
	topicStore := topicstore.NewSQLiteStore(db)
	commentStore := commentstore.NewSQLiteStore(db)

	uuidProvider := uuid.NewProvider()
	bcryptProvider := bcrypt.NewProvider()

	fc := &followCheckerAdapter{store: followStore}

	return usertransport.NewHandler(
		&authUserExtractor{},
		usercommands.NewRegisterHandler(userStore, uuidProvider, bcryptProvider),
		usercommands.NewLoginHandler(userStore, bcryptProvider, sessionMgr),
		usercommands.NewLogoutHandler(sessionMgr),
		usercommands.NewUpdateProfileHandler(userStore),
		usercommands.NewTogglePrivacyHandler(userStore),
		userqueries.NewGetProfileResolver(userStore, fc, followStore),
		userqueries.NewGetActivityResolver(userStore, topicStore, commentStore, topicStore, followStore),
		userqueries.NewListUsersResolver(userStore),
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

type coreSessionAdapter struct {
	inner sessionManager
}

func (a *coreSessionAdapter) Create(ctx context.Context, userID string) (*coresession.Session, error) {
	sess, err := a.inner.CreateSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &coresession.Session{
		Token:     sess.AccessToken,
		UserID:    sess.UserID,
		ExpiresAt: sess.Expiry,
	}, nil
}

func (a *coreSessionAdapter) Get(_ context.Context, token string) (*coresession.Session, error) {
	sess, err := a.inner.GetSession(token)
	if err != nil {
		return nil, err
	}
	return &coresession.Session{
		Token:     sess.AccessToken,
		UserID:    sess.UserID,
		ExpiresAt: sess.Expiry,
	}, nil
}

func (a *coreSessionAdapter) Revoke(_ context.Context, token string) error {
	return a.inner.DeleteSession(token)
}

var _ coresession.Manager = (*coreSessionAdapter)(nil)
