package eventbus

const (
	EventPost         = "post"
	EventPostLiked    = "post.liked"
	EventPostDisliked = "post.disliked"

	EventComment         = "comment"
	EventCommentLiked    = "comment.liked"
	EventCommentDisliked = "comment.disliked"

	EventFollowRequested = "follow.requested"
	EventFollowAccepted  = "follow.accepted"
	EventFollowDeclined  = "follow.declined"
	EventFollowRemoved   = "follow.removed"

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
