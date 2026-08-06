package group

import (
	"context"
	"errors"
	"time"
)

var (
	ErrGroupNotFound       = errors.New("group not found")
	ErrNotMember           = errors.New("user is not a member of this group")
	ErrAlreadyMember       = errors.New("user is already a member of this group")
	ErrAlreadyInvited      = errors.New("user is already invited to this group")
	ErrAlreadyRequested    = errors.New("user already has a pending join request")
	ErrNotAdmin            = errors.New("user is not a group admin or creator")
	ErrNotCreator          = errors.New("user is not the group creator")
	ErrSelfInvite          = errors.New("cannot invite yourself")
	ErrSelfJoin            = errors.New("cannot join your own group")
	ErrInvitationNotFound  = errors.New("invitation not found")
	ErrJoinRequestNotFound = errors.New("join request not found")
	ErrPostNotFound        = errors.New("group post not found")
	ErrInvalidVoteValue    = errors.New("reaction_type must be 1 (like) or -1 (dislike)")
)

type Role string

const (
	RoleCreator Role = "creator"
	RoleAdmin   Role = "admin"
	RoleMember  Role = "member"
)

type Group struct {
	ID               string
	Title            string
	Description      string
	CreatorID        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	MembershipStatus string
}

type Member struct {
	GroupID  string
	UserID   string
	Role     Role
	JoinedAt time.Time
}

type Invitation struct {
	ID        string
	GroupID   string
	InviterID string
	InviteeID string
	CreatedAt time.Time
}

type JoinRequest struct {
	ID          string
	GroupID     string
	RequesterID string
	CreatedAt   time.Time
}

type Post struct {
	ID            string
	GroupID       string
	AuthorID      string
	Title         string
	Content       string
	ImagePath     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	UpvoteCount   int
	DownvoteCount int
	VoteScore     int
	UserVote      *int
	CommentsCount int
}

type VoteCounts struct {
	Upvotes   int
	Downvotes int
	Score     int
}

type PostComment struct {
	ID        string
	PostID    string
	AuthorID  string
	Content   string
	ImagePath string
	CreatedAt time.Time
}

type Repository interface {
	Repo
	MemberRepository
	InvitationRepository
	JoinRequestRepository
	PostRepository
	PostCommentRepository
}

type Repo interface {
	CreateGroup(ctx context.Context, g *Group) error
	GetGroupByID(ctx context.Context, groupID string) (*Group, error)
	ListGroups(ctx context.Context, page, size int) ([]Group, int, error)
	UpdateGroup(ctx context.Context, id, title, description string) error
	DeleteGroup(ctx context.Context, id string) error
}

type MemberRepository interface {
	AddMember(ctx context.Context, groupID, userID string, role Role) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
	GetMemberRole(ctx context.Context, groupID, userID string) (Role, error)
	CountMembers(ctx context.Context, groupID string) (int, error)
	GetGroupMembers(ctx context.Context, groupID string, page, size int) ([]Member, int, error)
}

type InvitationRepository interface {
	CreateInvitation(ctx context.Context, inv *Invitation) error
	DeleteInvitation(ctx context.Context, groupID, inviteeID string) error
	GetInvitation(ctx context.Context, groupID, inviteeID string) (*Invitation, error)
	IsInvited(ctx context.Context, groupID, inviteeID string) (bool, error)
	GetPendingInvitations(ctx context.Context, userID string) ([]Invitation, error)
}

type JoinRequestRepository interface {
	CreateJoinRequest(ctx context.Context, jr *JoinRequest) error
	DeleteJoinRequest(ctx context.Context, groupID, requesterID string) error
	GetJoinRequest(ctx context.Context, groupID, requesterID string) (*JoinRequest, error)
	GetJoinRequestByID(ctx context.Context, id string) (*JoinRequest, error)
	HasPendingRequest(ctx context.Context, groupID, requesterID string) (bool, error)
	GetPendingJoinRequests(ctx context.Context, groupID string) ([]JoinRequest, error)
}

type PostRepository interface {
	CreatePost(ctx context.Context, p *Post) error
	GetPostsByGroupID(ctx context.Context, groupID, userID string, page, size int) ([]Post, int, error)
	CastPostVote(ctx context.Context, userID, postID string, reactionType int) error
	GetPostVoteCounts(ctx context.Context, postID string) (*VoteCounts, error)
}

type PostCommentRepository interface {
	CreatePostComment(ctx context.Context, c *PostComment) error
	GetPostComments(ctx context.Context, postID string, page, size int) ([]PostComment, int, error)
	CountPostComments(ctx context.Context, postID string) (int, error)
}

type FollowChecker interface {
	AreConnected(ctx context.Context, a, b string) (bool, error)
}

type EventBus interface {
	Publish(ctx context.Context, eventType string, payload any) error
}

type ImageStorage interface {
	Upload(ctx context.Context, data []byte, path string) error
}
