package eventbus

const (
	EventPost         = "post"
	EventPostLiked    = "post.liked"
	EventPostDisliked = "post.disliked"

	EventComment         = "comment"
	EventCommentLiked    = "comment.liked"
	EventCommentDisliked = "comment.disliked"

	EventPostVoteDeleted    = "post.vote.deleted"
	EventCommentVoteDeleted = "comment.vote.deleted"

	EventFollow          = "follow"
	EventFollowRequested = "follow.requested"
	EventFollowAccepted  = "follow.accepted"
	EventFollowDeclined  = "follow.declined"

	EventGroup              = "group"
	EventGroupInvitation    = "group.invitation"
	EventGroupJoinRequested = "group.join.requested"
	EventGroupJoinAccepted  = "group.join.accepted"
	EventGroupJoinDeclined  = "group.join.declined"

	EventGroupInviteAccepted        = "group.invitation.accepted"
	EventGroupInviteAcceptedPending = "group.invitation.accepted_pending"
	EventGroupInviteDeclined        = "group.invitation.declined"

	EventProfileUpdate = "profile.updated"

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
	Type               string   `json:"type"`
	RecipientID        string   `json:"recipient_id"`
	ActorID            string   `json:"actor_id"`
	ActorName          string   `json:"actor_name"`
	ActorAvatar        string   `json:"actor_avatar"`
	ResourceType       string   `json:"resource_type"`
	ResourceID         string   `json:"resource_id"`
	GroupID            string   `json:"group_id"`
	JoinRequestID      string   `json:"join_request_id"`
	MultipleRecipients []string `json:"multiple_recipients"`
	ContentText        string   `json:"content_text"`
	ImageURL           string   `json:"image_url"`
	EventID            string   `json:"event_id"`
}
