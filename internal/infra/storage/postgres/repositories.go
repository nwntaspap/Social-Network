package postgres

import (
	"database/sql"

	"github.com/arnald/forum/internal/domain/activity"
	"github.com/arnald/forum/internal/domain/category"
	"github.com/arnald/forum/internal/domain/chat"
	"github.com/arnald/forum/internal/domain/comment"
	"github.com/arnald/forum/internal/domain/notification"
	"github.com/arnald/forum/internal/domain/oauth"
	"github.com/arnald/forum/internal/domain/topic"
	"github.com/arnald/forum/internal/domain/user"
	"github.com/arnald/forum/internal/domain/vote"
	activities "github.com/arnald/forum/internal/infra/storage/postgres/activity"
	"github.com/arnald/forum/internal/infra/storage/postgres/categories"
	"github.com/arnald/forum/internal/infra/storage/postgres/chats"
	"github.com/arnald/forum/internal/infra/storage/postgres/comments"
	"github.com/arnald/forum/internal/infra/storage/postgres/notifications"
	oauthrepo "github.com/arnald/forum/internal/infra/storage/postgres/oauth"
	"github.com/arnald/forum/internal/infra/storage/postgres/topics"
	"github.com/arnald/forum/internal/infra/storage/postgres/users"
	"github.com/arnald/forum/internal/infra/storage/postgres/votes"
)

type Repositories struct {
	UserRepo         user.Repository
	CategoryRepo     category.Repository
	TopicRepo        topic.Repository
	CommentRepo      comment.Repository
	VoteRepo         vote.Repository
	NotificationRepo notification.Repository
	OauthRepo        oauth.Repository
	ActivityRepo     activity.Repository
	ChatRepo         chat.Repository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		UserRepo:         users.NewRepo(db),
		CategoryRepo:     categories.NewRepo(db),
		TopicRepo:        topics.NewRepo(db),
		CommentRepo:      comments.NewRepo(db),
		VoteRepo:         votes.NewRepo(db),
		OauthRepo:        oauthrepo.NewOAuthRepository(db),
		ActivityRepo:     activities.NewRepo(db),
		ChatRepo:         chats.NewRepo(db),
		NotificationRepo: notifications.NewRepo(db),
	}
}
