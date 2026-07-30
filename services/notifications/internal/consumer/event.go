package consumer

import (
	"social-network/services/notifications/internal/platform/eventbus"
	"social-network/services/notifications/internal/store"
)

type EventEnvelope struct {
	Type         string `json:"type"`
	RecipientID  string `json:"recipient_id"`
	ActorID      string `json:"actor_id"`
	ActorName    string `json:"actor_name"`
	ActorAvatar  string `json:"actor_avatar"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
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
	case eventbus.EventPostLiked:
		return "like"
	case eventbus.EventPostDisliked:
		return "dislike"
	case eventbus.EventCommentLiked:
		return "like"
	case eventbus.EventCommentDisliked:
		return "dislike"
	case eventbus.EventFollowRequested:
		return "follow_request"
	case "follow.requested.deleted":
		return "follow_cancelled"
	case eventbus.EventFollowAccepted:
		return "follow_accept"
	case "follow.accepted.deleted":
		return "unfollow"
	case eventbus.EventFollowDeclined:
		return "follow_declined"
	case eventbus.EventFollowRemoved:
		return "follow_removed"
	case eventbus.EventGroupInvitation:
		return "group_invite"
	case "group.invitation.deleted":
		return "group_invite_removed"
	case eventbus.EventGroupJoinRequested:
		return "group_join_request"
	default:
		return eventType
	}
}
