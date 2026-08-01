package queries

import (
	"context"
	"time"

	"social-network/internal/chat"
)

type GetChatUsersRequest struct {
	MeID string
}

type Conversation struct {
	ID            string
	OtherUser     chat.UserRef
	UnreadCount   int
	IsOnline      bool
	LastMessageAt *time.Time
	CreatedAt     time.Time
}

type GetChatUsersResolver struct {
	chatRepo    chat.Repository
	userRepo    chat.UserRepository
	broadcaster chat.Broadcaster
}

func NewGetChatUsersResolver(chatRepo chat.Repository, userRepo chat.UserRepository, broadcaster chat.Broadcaster) *GetChatUsersResolver {
	return &GetChatUsersResolver{
		chatRepo:    chatRepo,
		userRepo:    userRepo,
		broadcaster: broadcaster,
	}
}

func (r *GetChatUsersResolver) Resolve(ctx context.Context, req GetChatUsersRequest) ([]Conversation, error) {
	allUsers, err := r.userRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	usersByID := make(map[string]chat.UserRef, len(allUsers))
	for _, u := range allUsers {
		usersByID[u.ID] = *u
	}

	myChats, err := r.chatRepo.GetChatsForUser(ctx, req.MeID)
	if err != nil {
		return nil, err
	}

	conversations := make([]Conversation, 0, len(myChats))
	for _, c := range myChats {
		otherID := c.UserTwoID
		if otherID == req.MeID {
			otherID = c.UserOneID
		}

		other, ok := usersByID[otherID]
		if !ok {
			other = chat.UserRef{ID: otherID}
		}

		conversations = append(conversations, Conversation{
			ID:            c.ID,
			OtherUser:     other,
			UnreadCount:   c.UnreadCount,
			IsOnline:      r.broadcaster.IsOnline(otherID),
			LastMessageAt: c.LastMessageAt,
			CreatedAt:     c.CreatedAt,
		})
	}

	return conversations, nil
}
