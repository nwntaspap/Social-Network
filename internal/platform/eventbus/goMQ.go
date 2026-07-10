package eventbus

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/danielkotsi/golangMQSDK/gomqSDK"
	"github.com/danielkotsi/golangMQSDK/protocol"
)

//go:embed gobroker_schema.json
var goBrokerSchema []byte

type GoBroker struct {
	client *gomqSDK.Client
	// this is a channel used for all the publishes
	pubChannel *gomqSDK.ClientChannel
	// becasue this channel is used everywhere it need a lock
	pubMu sync.Mutex
}
type goBrokerMessage struct {
	body        []byte
	routingkey  string
	exchange    string
	deliveryTag uint16
	channel     *gomqSDK.ClientChannel
}

func (m *goBrokerMessage) Body() []byte {
	return m.body
}

func (m *goBrokerMessage) Exchange() string {
	return m.exchange
}

func (m *goBrokerMessage) RoutingKey() string {
	return m.routingkey
}

func (m *goBrokerMessage) Ack() error {
	return m.channel.Ack(m.deliveryTag)
}

func (m *goBrokerMessage) Nack(requeue bool) error {
	return m.channel.Nack(m.deliveryTag, requeue)
}

func NewGoBroker() (*GoBroker, error) {
	cfg := gomqSDK.Config{
		ClientName:   "social-network",
		Username:     "social-network",
		Password:     "123456789",
		ChannelMax:   10,
		FrameMax:     10372,
		HeartbeatSec: 10,
	}
	client, err := gomqSDK.Connect("localhost:5672", cfg)
	if err != nil {
		return &GoBroker{}, fmt.Errorf("not able to connect to broker:%w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	channel, err := client.OpenChannel(ctx)
	if err != nil {
		return &GoBroker{}, fmt.Errorf("channel for communication could not be created:%w", err)
	}
	broker := &GoBroker{
		client:     client,
		pubChannel: channel,
	}
	err = broker.InitTopology(ctx)
	if err != nil {
		return &GoBroker{}, fmt.Errorf("decleration error:%w", err)
	}

	return broker, nil
}

func (b *GoBroker) Subscribe(ctx context.Context, queue string) (<-chan Message, error) {
	channel, err := b.client.OpenChannel(ctx)
	if err != nil {
		return nil, fmt.Errorf("channel for communication could not be created:%w", err)
	}
	incoming, err := channel.Consume(queue, ctx)
	if err != nil {
		return nil, fmt.Errorf("consume error:%w", err)
	}

	// this is a bideriectional channel
	outgoing := make(chan Message, 100)
	go b.consume(incoming, outgoing, channel)

	// a recieve only channel is returned
	return outgoing, nil
}

// this one just sends from the provider specific chan to our generic interface chan
func (b *GoBroker) consume(incoming chan protocol.Deliver, outgoing chan Message, ch *gomqSDK.ClientChannel) {
	defer close(outgoing)
	for msg := range incoming {
		event := goBrokerMessage{
			body:        msg.Body,
			routingkey:  msg.RoutingKey,
			deliveryTag: msg.DeliveryTag,
			exchange:    msg.Exchange,
			channel:     ch,
		}
		outgoing <- &event
	}
}

type goBrokerTopology struct {
	Exchanges []exchangeDecl `json:"exchanges"`
	Queues    []queueDecl    `json:"queues"`
	Bindings  []bindingDecl  `json:"bindings"`
}

type exchangeDecl struct {
	Name string `json:"name"`
}

type queueDecl struct {
	Name                 string `json:"name"`
	DeadLetterExchange   string `json:"dead_letter_exchange"`
	DeadLetterRoutingKey string `json:"dead_letter_routing_key"`
}

type bindingDecl struct {
	Queue      string `json:"queue"`
	Exchange   string `json:"exchange"`
	RoutingKey string `json:"routing_key"`
}

func (b *GoBroker) InitTopology(ctx context.Context) error {
	var topo goBrokerTopology
	if err := json.Unmarshal(goBrokerSchema, &topo); err != nil {
		return fmt.Errorf("parse topology: %w", err)
	}

	for _, ex := range topo.Exchanges {
		if _, err := b.pubChannel.DeclareExchange(ex.Name, ctx); err != nil {
			return fmt.Errorf("declare exchange %s: %w", ex.Name, err)
		}
	}
	for _, q := range topo.Queues {
		if _, err := b.pubChannel.DeclareQueue(q.Name, ctx, q.DeadLetterExchange, q.DeadLetterRoutingKey); err != nil {
			return fmt.Errorf("declare queue %s: %w", q.Name, err)
		}
	}
	for _, bnd := range topo.Bindings {
		if err := b.pubChannel.BindQueue(bnd.Queue, bnd.Exchange, bnd.RoutingKey, ctx); err != nil {
			return fmt.Errorf("bind %s: %w", bnd.Queue, err)
		}
	}
	return nil
}

func (b *GoBroker) Publish(exchange, routingkey string, body []byte) error {
	msg := protocol.Publish{
		Exchange:   exchange,
		RoutingKey: routingkey,
		Body:       body,
	}
	b.pubMu.Lock()
	defer b.pubMu.Unlock()

	return b.pubChannel.Publish(msg)
}
