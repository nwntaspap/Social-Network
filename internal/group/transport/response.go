package transport

import (
	"time"

	"social-network/internal/group"
	"social-network/internal/group/queries"
)

type UserResult struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Nickname    string `json:"nickname,omitempty"`
	AboutMe     string `json:"aboutMe,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	DateOfBirth string `json:"dateOfBirth"`
	IsPublic    bool   `json:"isPublic"`
	CreatedAt   string `json:"createdAt"`
}

type GroupResponse struct {
	ID               string      `json:"id"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	CreatorID        string      `json:"creatorId"`
	Creator          *UserResult `json:"creator"`
	MembersCount     int         `json:"membersCount"`
	MembershipStatus string      `json:"membershipStatus"`
	UnreadCount      int         `json:"unreadCount"`
	CreatedAt        string      `json:"createdAt"`
	UpdatedAt        string      `json:"updatedAt"`
}

type GroupDetailResponse struct {
	ID               string      `json:"id"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	CreatorID        string      `json:"creatorId"`
	Creator          *UserResult `json:"creator"`
	MembersCount     int         `json:"membersCount"`
	MembershipStatus string      `json:"membershipStatus"`
	CreatedAt        string      `json:"createdAt"`
	UpdatedAt        string      `json:"updatedAt"`
}

type GroupPostResponse struct {
	ID            string      `json:"id"`
	GroupID       string      `json:"groupId"`
	UserID        string      `json:"userId"`
	User          *UserResult `json:"user"`
	Title         string      `json:"title,omitempty"`
	Content       string      `json:"content"`
	ImageURL      string      `json:"imageUrl,omitempty"`
	Privacy       string      `json:"privacy"`
	CommentsCount int         `json:"commentsCount"`
	LikesCount    int         `json:"likesCount"`
	DislikesCount int         `json:"dislikesCount"`
	IsLiked       bool        `json:"isLiked"`
	UserVote      *int        `json:"userVote"`
	CreatedAt     string      `json:"createdAt"`
	UpdatedAt     string      `json:"updatedAt"`
}

type GroupPostCommentResponse struct {
	ID        string      `json:"id"`
	PostID    string      `json:"postId"`
	UserID    string      `json:"userId"`
	User      *UserResult `json:"user"`
	Content   string      `json:"content"`
	ImageURL  string      `json:"imageUrl,omitempty"`
	CreatedAt string      `json:"createdAt"`
}

type InvitationResponse struct {
	ID        string      `json:"id"`
	GroupID   string      `json:"groupId"`
	Group     *GroupBrief `json:"group,omitempty"`
	InviterID string      `json:"inviterId"`
	Inviter   *UserResult `json:"inviter"`
	InviteeID string      `json:"inviteeId"`
	Invitee   *UserResult `json:"invitee"`
	CreatedAt string      `json:"createdAt"`
}

type JoinRequestResponse struct {
	ID          string      `json:"id"`
	GroupID     string      `json:"groupId"`
	Group       *GroupBrief `json:"group"`
	RequesterID string      `json:"requesterId"`
	Requester   *UserResult `json:"requester"`
	CreatedAt   string      `json:"createdAt"`
}

type GroupMemberResponse struct {
	GroupID  string      `json:"groupId"`
	UserID   string      `json:"userId"`
	User     *UserResult `json:"user"`
	Role     string      `json:"role"`
	JoinedAt string      `json:"joinedAt"`
}

type GroupBrief struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func toGroupResponse(g *group.Group, creator *UserResult, membersCount int, membershipStatus string) GroupResponse {
	return GroupResponse{
		ID:               g.ID,
		Title:            g.Title,
		Description:      g.Description,
		CreatorID:        g.CreatorID,
		Creator:          creator,
		MembersCount:     membersCount,
		MembershipStatus: membershipStatus,
		UnreadCount:      g.UnreadCount,
		CreatedAt:        formatTime(g.CreatedAt),
		UpdatedAt:        formatTime(g.UpdatedAt),
	}
}

func toGroupDetailResponse(res *queries.GetGroupResult, creator *UserResult) GroupDetailResponse {
	return GroupDetailResponse{
		ID:               res.Group.ID,
		Title:            res.Group.Title,
		Description:      res.Group.Description,
		CreatorID:        res.Group.CreatorID,
		Creator:          creator,
		MembersCount:     res.MembersCount,
		MembershipStatus: res.MembershipStatus,
		CreatedAt:        formatTime(res.Group.CreatedAt),
		UpdatedAt:        formatTime(res.Group.UpdatedAt),
	}
}

func toGroupPostResponse(p *group.Post, user *UserResult) GroupPostResponse {
	isLiked := p.UserVote != nil && *p.UserVote == 1
	return GroupPostResponse{
		ID:            p.ID,
		GroupID:       p.GroupID,
		UserID:        p.AuthorID,
		User:          user,
		Title:         p.Title,
		Content:       p.Content,
		ImageURL:      p.ImagePath,
		Privacy:       "public",
		CommentsCount: p.CommentsCount,
		LikesCount:    p.UpvoteCount,
		DislikesCount: p.DownvoteCount,
		IsLiked:       isLiked,
		UserVote:      p.UserVote,
		CreatedAt:     formatTime(p.CreatedAt),
		UpdatedAt:     formatTime(p.UpdatedAt),
	}
}

func toGroupPostCommentResponse(c *group.PostComment, user *UserResult) GroupPostCommentResponse {
	return GroupPostCommentResponse{
		ID:        c.ID,
		PostID:    c.PostID,
		UserID:    c.AuthorID,
		User:      user,
		Content:   c.Content,
		ImageURL:  c.ImagePath,
		CreatedAt: formatTime(c.CreatedAt),
	}
}

func toInvitationResponse(inv *group.Invitation, inviter, invitee *UserResult, group *GroupBrief) InvitationResponse {
	return InvitationResponse{
		ID:        inv.ID,
		GroupID:   inv.GroupID,
		Group:     group,
		InviterID: inv.InviterID,
		Inviter:   inviter,
		InviteeID: inv.InviteeID,
		Invitee:   invitee,
		CreatedAt: formatTime(inv.CreatedAt),
	}
}

func toJoinRequestResponse(jr *group.JoinRequest, groupBrief *GroupBrief, requester *UserResult) JoinRequestResponse {
	return JoinRequestResponse{
		ID:          jr.ID,
		GroupID:     jr.GroupID,
		Group:       groupBrief,
		RequesterID: jr.RequesterID,
		Requester:   requester,
		CreatedAt:   formatTime(jr.CreatedAt),
	}
}

func toGroupMemberResponse(m *group.Member, user *UserResult) GroupMemberResponse {
	return GroupMemberResponse{
		GroupID:  m.GroupID,
		UserID:   m.UserID,
		User:     user,
		Role:     string(m.Role),
		JoinedAt: formatTime(m.JoinedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
