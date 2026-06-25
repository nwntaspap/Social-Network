package sqlite

import (
	"database/sql"

	"social-network/internal/domain/activity"
	"social-network/internal/domain/category"
	"social-network/internal/domain/chat"
	"social-network/internal/domain/comment"
	"social-network/internal/domain/notification"
	"social-network/internal/domain/oauth"
	"social-network/internal/domain/topic"
	"social-network/internal/domain/user"
	"social-network/internal/domain/vote"
	"social-network/internal/infra/storage/sqlite/categories"
	"social-network/internal/infra/storage/sqlite/chats"
	"social-network/internal/infra/storage/sqlite/comments"
	"social-network/internal/infra/storage/sqlite/notifications"
	"social-network/internal/infra/storage/sqlite/topics"
	"social-network/internal/infra/storage/sqlite/users"
	"social-network/internal/infra/storage/sqlite/votes"

	activities "social-network/internal/infra/storage/sqlite/activity"

	oauthrepo "social-network/internal/infra/storage/sqlite/oauth"
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
