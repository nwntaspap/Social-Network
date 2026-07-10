package eventbus

import "context"

type EventBus interface {
	Subscribe(queue string, ctx context.Context) (incoming <-chan Message, err error)
	Publish(exchange, routingkey string, body []byte) error
	InitTopology(ctx context.Context) error
}

// in order to make this interface generic and to work with any kind of broker
// we also need to have an interface for the messages
// because different provider needs different actions
// i.e. rabbitMQ ACK/NACK
//     KAFKA consumer.seek or smth
// so for now this interface will be used for the message

// so this covers our own broker and rabbitMQ as it is
// for other providers the behaviour is going to change inside the implementation
type Message interface {
	Body() []byte
	RoutingKey() string
	Ack() error
	// requeue true in order to requeue
	// false in order to go to dead letter que
	Nack(requeue bool) error
}
