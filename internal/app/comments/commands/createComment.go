package commentcommands

import (
	"context"
	"fmt"
	"strconv"

	"social-network/internal/app/notifications"
	"social-network/internal/domain/comment"
	"social-network/internal/domain/notification"
	"social-network/internal/domain/topic"
	"social-network/internal/domain/user"
)

type CreateCommentRequest struct {
	User    *user.User
	Content string `json:"content"`
	TopicID int    `json:"topicId"`
}

type CreateCommentRequestHandler interface {
	Handle(ctx context.Context, req CreateCommentRequest) (*comment.Comment, error)
}

type createCommentRequestHandler struct {
	commentrepo      comment.Repository
	topicrepo        topic.Repository
	notificationrepo notification.Repository
	notifier         notifications.Notifier
}

func NewCreateCommentRequestHandler(commentrepo comment.Repository, topicrepo topic.Repository, notificationrepo notification.Repository, notifier notifications.Notifier) CreateCommentRequestHandler {
	return &createCommentRequestHandler{
		commentrepo:      commentrepo,
		topicrepo:        topicrepo,
		notificationrepo: notificationrepo,
		notifier:         notifier,
	}
}

func (h *createCommentRequestHandler) Handle(ctx context.Context, req CreateCommentRequest) (*comment.Comment, error) {
	comment := &comment.Comment{
		UserID:  req.User.ID,
		TopicID: req.TopicID,
		Content: req.Content,
	}

	// only issue here is in the instance of a successful comment write in the db and a failure in a notification write it returns an err and the handler doesnt know what happened, could be fixed with error types
	err := h.commentrepo.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}
	topic, err := h.topicrepo.GetTopicByID(ctx, req.TopicID, &req.User.ID)
	if err != nil {
		return nil, err
	}
	if req.User.ID != topic.UserID {
		notification := &notification.Notification{
			ActorID:     req.User.Nickname,
			UserID:      topic.UserID,
			RelatedID:   strconv.Itoa(comment.TopicID),
			RelatedType: "topic",
			Type:        notification.NotificationTypeReply,
			Title:       "New comment",
			Message:     fmt.Sprintf("%s commented on your Topic %s", req.User.Nickname, topic.Title),
		}
		err = h.notificationrepo.Create(ctx, notification)
		if err != nil {
			return nil, err
		}
		h.notifier.BroadcastToUser(notification.UserID, notification)
	}

	return comment, nil
}
