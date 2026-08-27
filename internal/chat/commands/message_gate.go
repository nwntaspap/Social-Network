package commands

import (
	"context"
	"errors"

	"social-network/internal/chat"
)

var (
	// ErrNotConnected is returned when neither user follows the other.
	ErrNotConnected = errors.New("users are not connected: at least one must follow the other")

	// ErrCannotMessage is returned when the recipient neither follows the sender
	// nor has a public profile, so the message cannot be delivered to them.
	ErrCannotMessage = errors.New("recipient cannot receive your message")
)

// MessageGate enforces the private-messaging rules:
//   - the sender and recipient must be connected (at least one follows the other), and
//   - the recipient must be able to receive the message (they follow the sender
//     or their profile is public).
type MessageGate struct {
	follow  chat.FollowChecker
	privacy chat.UserPrivacyChecker
}

func NewMessageGate(follow chat.FollowChecker, privacy chat.UserPrivacyChecker) *MessageGate {
	return &MessageGate{follow: follow, privacy: privacy}
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

	isPrivate, err := g.privacy.IsPrivate(ctx, receiverID)
	if err != nil {
		return err
	}
	if !receiverFollows && isPrivate {
		return ErrCannotMessage
	}

	return nil
}
