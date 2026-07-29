package eventbus

import "context"

type Envelope struct {
	Type         string `json:"type"`
	RecipientID  string `json:"recipient_id"`
	ActorID      string `json:"actor_id"`
	ActorName    string `json:"actor_name"`
	ActorAvatar  string `json:"actor_avatar"`
	ResourceType string `json:"resource_type"`
	ResourceID   int    `json:"resource_id"`
	ContentText  string `json:"content_text"`
	ImageURL     string `json:"image_url"`
}

type EventBus interface {
	Subscribe(ctx context.Context, queue string) (incoming <-chan Message, err error)
	Publish(exchange, routingkey string, body []byte) error
	InitTopology(ctx context.Context) error
}

// in order to make this interface generic and to work with any kind of broker
// we also need to have an interface for the messages
// because different provider needs different actions
// i.e. rabbitMQ ACK/NACK
//     KAFKA consumer.seek or smth
// so for now this interface will be used for the message

// Message: so this covers our own broker and rabbitMQ as it is
// for other providers the behaviour is going to change inside the implementation
type Message interface {
	Body() []byte
	RoutingKey() string
	Ack() error
	// requeue true in order to requeue
	// false in order to go to dead letter que
	Nack(requeue bool) error
}
