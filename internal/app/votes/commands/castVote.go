package votecommands

import (
	"context"
	"fmt"
	"strconv"

	"social-network/internal/app/notifications"
	"social-network/internal/domain/comment"
	"social-network/internal/domain/notification"
	"social-network/internal/domain/topic"
	"social-network/internal/domain/vote"
)

type CastVoteRequest struct {
	Target       vote.Target `json:"target"`
	UserID       string      `json:"userId"`
	NickName     string
	ReactionType int `json:"reactionType"`
}

type castVoteRequestHandler struct {
	VoteRepo         vote.Repository
	TopicRepo        topic.Repository
	CommentRepo      comment.Repository
	NotificationRepo notification.Repository
	Notifier         notifications.Notifier
}

type CastVoteRequestHandler interface {
	Handle(ctx context.Context, req CastVoteRequest) error
}

func NewCastVoteHandler(voteRepo vote.Repository, topicRepo topic.Repository, commentRepo comment.Repository, notificationRepo notification.Repository, notifier notifications.Notifier) CastVoteRequestHandler {
	return &castVoteRequestHandler{
		VoteRepo:         voteRepo,
		TopicRepo:        topicRepo,
		CommentRepo:      commentRepo,
		NotificationRepo: notificationRepo,
		Notifier:         notifier,
	}
}

func (h *castVoteRequestHandler) Handle(ctx context.Context, req CastVoteRequest) error {
	err := h.VoteRepo.CastVote(ctx, req.UserID, req.Target, req.ReactionType)
	if err != nil {
		return err
	}
	// notification send, if failed, fails silently, we could fix that
	h.sendVoteNotification(ctx, req)

	return nil
}

func (h *castVoteRequestHandler) sendVoteNotification(ctx context.Context, req CastVoteRequest) {
	var ownerID string
	var contentType string
	var contentID string

	switch {
	case req.Target.CommentID != nil:
		comment, err := h.CommentRepo.GetCommentByID(ctx, *req.Target.CommentID)
		if err != nil {
			return
		}
		ownerID = comment.UserID
		contentType = "comment"
		contentID = strconv.Itoa(*req.Target.CommentID)
	case req.Target.TopicID != nil:
		topic, err := h.TopicRepo.GetTopicByID(ctx, *req.Target.TopicID, nil)
		if err != nil {
			return
		}
		ownerID = topic.UserID
		contentType = "topic"
		contentID = strconv.Itoa(*req.Target.TopicID)
	}

	if ownerID == "" || ownerID == req.UserID {
		return
	}

	var message string
	var title string
	var notificationType notification.Type
	switch req.ReactionType {
	case 1:
		message = fmt.Sprintf("%s liked your %s", req.NickName, contentType)
		title = "New like!"
		notificationType = notification.NotificationTypeLike
	case -1:
		message = fmt.Sprintf("%s disliked your %s", req.NickName, contentType)
		title = "New dislike!"
		notificationType = notification.NotificationTypeDislike
	}

	notification := &notification.Notification{
		Type:        notificationType,
		RelatedID:   contentID,
		UserID:      ownerID,
		ActorID:     req.UserID,
		Message:     message,
		Title:       title,
		RelatedType: contentType,
	}

	err := h.NotificationRepo.Create(ctx, notification)
	if err != nil {
		return
	}

	h.Notifier.BroadcastToUser(notification.UserID, notification)
}
