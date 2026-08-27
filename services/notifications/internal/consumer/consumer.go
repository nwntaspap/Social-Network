package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"social-network/services/notifications/internal/handler"
	"social-network/services/notifications/internal/platform/eventbus"
	"social-network/services/notifications/internal/store"
)

type Consumer struct {
	bus  eventbus.EventBus
	repo store.Repository
	hub  *handler.StreamHub
	log  *slog.Logger
}

func New(bus eventbus.EventBus, repo store.Repository, hub *handler.StreamHub, log *slog.Logger) *Consumer {
	return &Consumer{bus: bus, repo: repo, hub: hub, log: log}
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.bus.Subscribe(ctx, "notifications_queue")
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			c.handle(msg)
		}
	}
}

func (c *Consumer) handle(msg eventbus.Message) {
	var env EventEnvelope

	if err := json.Unmarshal(msg.Body(), &env); err != nil {
		c.log.Warn("failed to decode event", "error", err, "body", string(msg.Body()))
		_ = msg.Nack(false)
		return
	}

	routingKey := msg.RoutingKey()

	switch routingKey {
	case eventbus.RoutingCreated:
		c.handleCreated(env, msg)
	case eventbus.RoutingDeleted:
		c.handleDeleted(env, msg)
	case eventbus.RoutingUpdated:
		c.handleUpdated(env, msg)
	default:
		c.log.Warn("unknown routing key", "routing_key", routingKey)
		_ = msg.Nack(false)
	}
}

func (c *Consumer) handleCreated(env EventEnvelope, msg eventbus.Message) {
	if env.Type == eventbus.EventGroupJoinRequested {
		c.handleJoinRequest(env, msg)
		return
	}
	if env.Type == eventbus.EventEvent {
		c.handleEvent(env, msg)
		return
	}

	if env.RecipientID == "" || env.Type == "" || env.ActorID == "" {
		c.log.Warn("incomplete event", "envelope", env)
		_ = msg.Nack(false)
		return
	}

	switch env.Type {
	case eventbus.EventGroupJoinAccepted, eventbus.EventGroupJoinDeclined:
		if _, err := c.repo.DeleteByJoinRequestID(context.Background(), env.JoinRequestID); err != nil && !errors.Is(err, store.ErrNotFound) {
			c.log.Warn("failed to clear join request notifications", "error", err)
		}
	}

	notif := env.ToNotification()

	if err := c.repo.Create(context.Background(), notif); err != nil {
		c.log.Error("failed to create notification", "error", err)
		_ = msg.Nack(false)
		return
	}

	c.hub.Publish(env.RecipientID, *notif)

	if err := msg.Ack(); err != nil {
		c.log.Warn("failed to ack message", "error", err)
	}
}

func (c *Consumer) handleEvent(env EventEnvelope, msg eventbus.Message) {
	if env.Type == "" || env.ActorID == "" || len(env.MultipleRecipients) == 0 {
		c.log.Warn("incomplete event created", "envelope", env)
		_ = msg.Nack(false)
		return
	}

	for _, groupMember := range env.MultipleRecipients {
		notif := env.ToNotification()
		notif.RecipientID = groupMember
		if err := c.repo.Create(context.Background(), notif); err != nil {
			c.log.Error("failed to create join request notification", "recipient", groupMember, "error", err)
			_ = msg.Nack(false)
			return
		}
		c.hub.Publish(groupMember, *notif)
	}

	if err := msg.Ack(); err != nil {
		c.log.Warn("failed to ack message", "error", err)
	}
}

func (c *Consumer) handleJoinRequest(env EventEnvelope, msg eventbus.Message) {
	if env.Type == "" || env.ActorID == "" || env.JoinRequestID == "" || len(env.MultipleRecipients) == 0 {
		c.log.Warn("incomplete join request event", "envelope", env)
		_ = msg.Nack(false)
		return
	}

	for _, adminID := range env.MultipleRecipients {
		notif := env.ToNotification()
		notif.RecipientID = adminID
		if err := c.repo.Create(context.Background(), notif); err != nil {
			c.log.Error("failed to create join request notification", "recipient", adminID, "error", err)
			_ = msg.Nack(false)
			return
		}
		c.hub.Publish(adminID, *notif)
	}

	if err := msg.Ack(); err != nil {
		c.log.Warn("failed to ack message", "error", err)
	}
}

func (c *Consumer) handleDeleted(env EventEnvelope, msg eventbus.Message) {
	if env.Type == "" {
		c.log.Warn("incomplete event", "envelope", env)
		_ = msg.Nack(false)
		return
	}

	var (
		deleted []store.Notification
		err     error
	)

	switch env.Type {
	case eventbus.EventPost, eventbus.EventComment:
		deleted, err = c.repo.DeleteAllByResource(context.Background(), env.ResourceID)
	case eventbus.EventFollow:
		deleted, err = c.repo.DeleteFollowNotifications(context.Background(), env.RecipientID, env.ActorID)
	case eventbus.EventPostVoteDeleted, eventbus.EventCommentVoteDeleted:
		deleted, err = c.repo.DeleteVoteNotifications(context.Background(), env.ActorID, env.ResourceType, env.ResourceID)
	case eventbus.EventEvent:
		deleted, err = c.repo.DeleteEventByRecipient(context.Background(), mapEventType(env.Type), env.RecipientID, env.EventID)
	default:
		deleted, err = c.repo.DeleteByResource(context.Background(), mapEventType(env.Type), env.ActorID, env.ResourceType, env.ResourceID)
	}

	if err != nil && !errors.Is(err, store.ErrNotFound) {
		c.log.Warn("failed to delete notifications", "error", err)
	}

	for i := range deleted {
		notif := deleted[i]
		notif.Deleted = true
		c.hub.Publish(notif.RecipientID, notif)
	}

	if err := msg.Ack(); err != nil {
		c.log.Warn("failed to ack message", "error", err)
	}
}

func (c *Consumer) handleUpdated(env EventEnvelope, msg eventbus.Message) {
	if env.ActorID == "" {
		c.log.Warn("incomplete update event", "envelope", env)
		_ = msg.Nack(false)
		return
	}

	if err := c.repo.UpdateActorInfo(context.Background(), env.ActorID, env.ActorName, env.ActorAvatar); err != nil {
		c.log.Warn("failed to update actor info", "error", err)
	}

	if err := msg.Ack(); err != nil {
		c.log.Warn("failed to ack message", "error", err)
	}
}
