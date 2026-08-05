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
