package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestBrokerAddr_Default(t *testing.T) {
	t.Setenv("NOTIFICATIONS_BROKER_URL", "")
	if got := brokerAddr(); got != "localhost:5672" {
		t.Errorf("brokerAddr() = %q, want %q", got, "localhost:5672")
	}
}

func TestBrokerAddr_EnvOverride(t *testing.T) {
	t.Setenv("NOTIFICATIONS_BROKER_URL", "broker:5672")
	if got := brokerAddr(); got != "broker:5672" {
		t.Errorf("brokerAddr() = %q, want %q", got, "broker:5672")
	}
}

func TestGoBroker_PublishWithoutConnection(t *testing.T) {
	broker := &GoBroker{}
	err := broker.Publish("notifications.exchange", "created", []byte("x"))
	if err == nil {
		t.Fatal("Publish on unconnected broker expected error, got nil")
	}
}

func TestGoBroker_SubscribeWithoutConnection(t *testing.T) {
	broker := &GoBroker{}
	if _, err := broker.Subscribe(context.Background(), "notifications_queue"); !errors.Is(err, ErrNotConnected) {
		t.Errorf("Subscribe error = %v, want ErrNotConnected", err)
	}
}

func TestTopologyJSON_UnmarshalsCorrectly(t *testing.T) {
	var topo goBrokerTopology
	if err := json.Unmarshal(goBrokerSchema, &topo); err != nil {
		t.Fatalf("failed to unmarshal embedded topology: %v", err)
	}

	if len(topo.Exchanges) != 2 {
		t.Errorf("got %d exchanges, want 2", len(topo.Exchanges))
	}
	if len(topo.Queues) != 2 {
		t.Errorf("got %d queues, want 2", len(topo.Queues))
	}
	if len(topo.Bindings) != 4 {
		t.Errorf("got %d bindings, want 20", len(topo.Bindings))
	}

	exNames := map[string]bool{}
	for _, ex := range topo.Exchanges {
		exNames[ex.Name] = true
	}
	if !exNames["notifications.exchange"] {
		t.Error("missing exchange: notifications.exchange")
	}
	if !exNames["dead.letter.exchange"] {
		t.Error("missing exchange: dead.letter.exchange")
	}

	qNames := map[string]bool{}
	for _, q := range topo.Queues {
		qNames[q.Name] = true
	}
	if !qNames["notifications_queue"] {
		t.Error("missing queue: notifications_queue")
	}
	if !qNames["dead.letter_queue"] {
		t.Error("missing queue: dead.letter_queue")
	}
}

func TestGoBrokerMessage_BodyAndRoutingKey(t *testing.T) {
	msg := &goBrokerMessage{
		body:       []byte("test-payload"),
		routingkey: "test.routing.key",
		exchange:   "test.exchange",
	}

	if string(msg.Body()) != "test-payload" {
		t.Errorf("Body() = %q, want %q", msg.Body(), "test-payload")
	}
	if msg.RoutingKey() != "test.routing.key" {
		t.Errorf("RoutingKey() = %q, want %q", msg.RoutingKey(), "test.routing.key")
	}
}

func TestGoBroker_SubscribePublishRoundtrip(t *testing.T) {
	broker, err := NewGoBroker()
	if err != nil {
		t.Skipf("broker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msgs, err := broker.Subscribe(ctx, "notifications_queue")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// brief pause for subscription to register on the broker
	time.Sleep(300 * time.Millisecond)

	payload := []byte("hello from test")
	if err := broker.Publish("notifications.exchange", "post.liked", payload); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case msg := <-msgs:
		if string(msg.Body()) != string(payload) {
			t.Errorf("Body() = %q, want %q", msg.Body(), payload)
		}
		if msg.RoutingKey() != "post.liked" {
			t.Errorf("RoutingKey() = %q, want %q", msg.RoutingKey(), "post.liked")
		}
		if err := msg.Ack(); err != nil {
			t.Errorf("Ack failed: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for published message")
	}
}

func TestGoBroker_NackRequeuesMessage(t *testing.T) {
	broker, err := NewGoBroker()
	if err != nil {
		t.Skipf("broker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msgs, err := broker.Subscribe(ctx, "notifications_queue")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	payload := []byte("nack-me")
	if err := broker.Publish("notifications.exchange", "post.liked", payload); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case msg := <-msgs:
		if err := msg.Nack(true); err != nil {
			t.Errorf("Nack(true) failed: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for nack message")
	}

	// nack(true) requeues — message should be redelivered
	select {
	case msg := <-msgs:
		if string(msg.Body()) != string(payload) {
			t.Errorf("after Nack(true) got body %q, want %q", msg.Body(), payload)
		}
		_ = msg.Ack()
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for requeued message after Nack(true)")
	}
}
