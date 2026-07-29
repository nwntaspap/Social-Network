package consumer

import (
	"social-network/services/notifications/internal/store"
)

type EventEnvelope struct {
	Type         string `json:"type"`
	RecipientID  string `json:"recipient_id"`
	ActorID      string `json:"actor_id"`
	ActorName    string `json:"actor_name"`
	ActorAvatar  string `json:"actor_avatar"`
	ResourceType string `json:"resource_type"`
	ResourceID   int    `json:"resource_id"`
	ContentText  string `json:"content_text"`
	ImageURL     string `json:"image_url"`
}

func (e *EventEnvelope) ToNotification() *store.Notification {
	return &store.Notification{
		RecipientID:  e.RecipientID,
		Type:         mapEventType(e.Type),
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		ActorID:      e.ActorID,
		ActorName:    e.ActorName,
		ActorAvatar:  e.ActorAvatar,
		ContentText:  e.ContentText,
		ImageURL:     e.ImageURL,
	}
}

func mapEventType(eventType string) string {
	switch eventType {
	case "post.created":
		return "post_created"
	case "post.deleted":
		return "post_deleted"
	case "post.liked":
		return "like"
	case "post.liked.deleted":
		return "unlike"
	case "comment.created":
		return "comment_created"
	case "comment.liked":
		return "like"
	case "comment.liked.deleted":
		return "unlike"
	case "follow.requested":
		return "follow_request"
	case "follow.requested.deleted":
		return "follow_cancelled"
	case "follow":
		return "follow"
	case "follow.deleted":
		return "follow_deleted"
	case "follow.accepted":
		return "follow_accept"
	case "follow.accepted.deleted":
		return "unfollow"
	case "follow.declined":
		return "follow_declined"
	case "group.invitation":
		return "group_invite"
	case "group.invitation.deleted":
		return "group_invite_removed"
	case "group.join_requested":
		return "group_join_request"
	case "event.created":
		return "event_created"
	case "event.created.deleted":
		return "event_removed"
	default:
		return eventType
	}
}
