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
	case "post.liked":
		return "like"
	case "comment.voted":
		return "like"
	case "post.unliked":
		return "unlike"
	case "comment.vote.deleted":
		return "unlike"
	case "follow.requested":
		return "follow_request"
	case "follow.accepted":
		return "follow_accept"
	case "follow.declined":
		return "follow_decline"
	case "comment.created":
		return "comment"
	case "post.created":
		return "post"
	default:
		return eventType
	}
}
