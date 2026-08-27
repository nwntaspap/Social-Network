package eventbus

import "context"

type EventBus interface {
	Subscribe(ctx context.Context, queue string) (incoming <-chan Message, err error)
	Publish(exchange, routingkey string, body []byte) error
	InitTopology(ctx context.Context) error
}

type Message interface {
	Body() []byte
	RoutingKey() string
	Ack() error
	Nack(requeue bool) error
}
