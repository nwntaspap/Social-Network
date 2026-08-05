package bootstrap

import (
	"net/http"

	commentcommands "social-network/internal/comment/commands"
	commentqueries "social-network/internal/comment/queries"
	commentstore "social-network/internal/comment/store"
	commenttransport "social-network/internal/comment/transport"
	"social-network/internal/core/middleware"
	localstorage "social-network/internal/infra/storage/local"
	"social-network/internal/platform/database"
	"social-network/internal/platform/eventbus"
	topicstore "social-network/internal/topic/store"
	"social-network/internal/user"
	userstore "social-network/internal/user/store"
)

func initComment(db database.DB, bus eventbus.EventBus) *commenttransport.Handler {
	store := commentstore.NewSQLiteStore(db)
	img := localstorage.NewLocalStorage()
	users := userstore.NewSQLiteStore(db)
	topics := topicstore.NewSQLiteStore(db)

	extractUser := func(r *http.Request) (string, bool) {
		uid := middleware.GetUserIDFromContext(r)
		if uid == "" {
			return "", false
		}
		return uid, true
	}

	userLookup := &commentUserLookupAdapter{repo: userstore.NewSQLiteStore(db)}

	return commenttransport.NewHandler(
		extractUser,
		userLookup,
		commentcommands.NewCreateCommentHandler(store, bus, users, topics, img),
		commentcommands.NewUpdateCommentHandler(store),
		commentcommands.NewDeleteCommentHandler(store, bus, topics, users),
		commentcommands.NewCastCommentVoteHandler(store, bus, users),
		commentcommands.NewDeleteCommentVoteHandler(store, bus),
		commentqueries.NewGetCommentByIDResolver(store),
		commentqueries.NewGetCommentByIDWithVotesResolver(store),
		commentqueries.NewGetCommentsByTopicResolver(store),
		commentqueries.NewGetCommentsByTopicWithVotesResolver(store),
		commentqueries.NewGetVoteCountsResolver(store),
	)
}

type commentUserLookupAdapter struct {
	repo user.Repository
}

func (a *commentUserLookupAdapter) GetUserByID(ctx context.Context, id string) (*commenttransport.UserResult, error) {
	u, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &commenttransport.UserResult{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Nickname,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		AboutMe:   u.AboutMe,
		IsPublic:  !u.IsPrivate,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if !u.DateOfBirth.IsZero() {
		result.DateOfBirth = u.DateOfBirth.Format("2006-01-02")
	}

	if u.AvatarPath != "" {
		result.AvatarURL = u.AvatarPath
	}

	return result, nil
}

var _ commenttransport.UserLookup = (*commentUserLookupAdapter)(nil)
