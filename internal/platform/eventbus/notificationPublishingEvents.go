package eventbus

const (
	EventPost         = "post"
	EventPostLiked    = "post.liked"
	EventPostDisliked = "post.disliked"

	EventComment         = "comment"
	EventCommentLiked    = "comment.liked"
	EventCommentDisliked = "comment.disliked"

	EventFollow          = "follow"
	EventFollowRequested = "follow.requested"
	EventFollowAccepted  = "follow.accepted"
	EventFollowDeclined  = "follow.declined"

	EventGroupInvitation    = "group.invitation"
	EventGroupJoinRequested = "group.join_requested"

	EventEvent = "event"
)

const (
	RoutingCreated = "created"
	RoutingDeleted = "deleted"
	RoutingUpdated = "updated"
)

const (
	ResourcePost    = "post"
	ResourceComment = "comment"
	ResourceUser    = "user"
	ResourceGroup   = "group"
	ResourceEvent   = "event"
)

type Notification struct {
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
