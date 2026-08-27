package queries

import (
	"context"

	"social-network/internal/group"
)

type GetGroupPresenceQuery struct {
	GroupID string
	UserID  string
}

type GroupPresenceMember struct {
	ID       string `json:"id"`
	IsOnline bool   `json:"isOnline"`
}

type GetGroupPresenceResult struct {
	GroupID string
	Total   int
	Online  int
	Members []GroupPresenceMember
}

type GroupPresenceRepository interface {
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
	ListGroupMemberIDs(ctx context.Context, groupID string) ([]string, error)
}

// GroupPresenceOnlineChecker reports whether a user currently has a live
// connection. It is injected by the transport layer so this query stays
// decoupled from the realtime hub.
type GroupPresenceOnlineChecker func(userID string) bool

type GetGroupPresenceResolver struct {
	repo     GroupPresenceRepository
	isOnline GroupPresenceOnlineChecker
}

func NewGetGroupPresenceResolver(repo GroupPresenceRepository, isOnline GroupPresenceOnlineChecker) *GetGroupPresenceResolver {
	return &GetGroupPresenceResolver{repo: repo, isOnline: isOnline}
}

// Resolve returns the group's member count, how many members are currently
// online, and per-member online flags. Only group members may resolve it.
func (r *GetGroupPresenceResolver) Resolve(ctx context.Context, q GetGroupPresenceQuery) (*GetGroupPresenceResult, error) {
	isMember, err := r.repo.IsMember(ctx, q.GroupID, q.UserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, group.ErrNotMember
	}

	ids, err := r.repo.ListGroupMemberIDs(ctx, q.GroupID)
	if err != nil {
		return nil, err
	}

	result := &GetGroupPresenceResult{
		GroupID: q.GroupID,
		Total:   len(ids),
		Members: make([]GroupPresenceMember, 0, len(ids)),
	}
	for _, id := range ids {
		online := r.isOnline != nil && r.isOnline(id)
		if online {
			result.Online++
		}
		result.Members = append(result.Members, GroupPresenceMember{ID: id, IsOnline: online})
	}
	return result, nil
}
