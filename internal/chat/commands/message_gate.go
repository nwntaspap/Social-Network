package commands

import (
	"context"
	"errors"

	"social-network/internal/chat"
)

// ErrNotConnected is returned when neither user follows the other.
var ErrNotConnected = errors.New("users are not connected: at least one must follow the other")

// MessageGate enforces the private-messaging rules: the sender and recipient
// must be connected (at least one of them follows the other). A follow
// established in either direction lets both users exchange messages.
type MessageGate struct {
	follow chat.FollowChecker
}

func NewMessageGate(follow chat.FollowChecker) *MessageGate {
	return &MessageGate{follow: follow}
}

// Validate returns an error if a message from senderID to receiverID is not allowed.
func (g *MessageGate) Validate(ctx context.Context, senderID, receiverID string) error {
	senderFollows, err := g.follow.AreConnected(ctx, senderID, receiverID)
	if err != nil {
		return err
	}
	receiverFollows, err := g.follow.AreConnected(ctx, receiverID, senderID)
	if err != nil {
		return err
	}

	if !senderFollows && !receiverFollows {
		return ErrNotConnected
	}

	return nil
}
