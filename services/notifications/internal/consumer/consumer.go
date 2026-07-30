package consumer

import (
	"context"
	"encoding/json"
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
	if env.RecipientID == "" || env.Type == "" || env.ActorID == "" {
		c.log.Warn("incomplete event", "envelope", env)
		_ = msg.Nack(false)
		return
	}

	notif := env.ToNotification()

	if err := c.repo.Create(context.Background(), notif); err != nil {
		c.log.Error("failed to create notification", "error", err)
		_ = msg.Nack(true)
		return
	}

	c.hub.Publish(env.RecipientID, *notif)

	if err := msg.Ack(); err != nil {
		c.log.Warn("failed to ack message", "error", err)
	}
}

func (c *Consumer) handleDeleted(env EventEnvelope, msg eventbus.Message) {
	if env.RecipientID == "" || env.Type == "" || env.ActorID == "" {
		c.log.Warn("incomplete event", "envelope", env)
		_ = msg.Nack(false)
		return
	}

	notif := env.ToNotification()
	notif.Deleted = true

	c.hub.Publish(env.RecipientID, *notif)

	switch env.Type {
	case eventbus.EventPost, eventbus.EventComment:
		if err := c.repo.DeleteAllByResource(context.Background(), env.ResourceID); err != nil {
			c.log.Warn("failed to cascade delete notifications", "error", err)
		}
	default:
		if err := c.repo.DeleteByResource(context.Background(), env.ActorID, env.ResourceType, env.ResourceID); err != nil {
			c.log.Warn("failed to delete notification", "error", err)
		}
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
