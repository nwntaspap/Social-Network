package realtime

import (
	"encoding/json"
	"testing"
)

type recorderHandler struct {
	calls []Envelope
}

func (r *recorderHandler) Handle(_ *Client, env Envelope) {
	r.calls = append(r.calls, env)
}

func TestRouter_DispatchesByType(t *testing.T) {
	send := &recorderHandler{}
	ping := &recorderHandler{}
	router := NewWSRouter(map[string]WSHandler{
		TypeChatSend: send,
		TypePing:     ping,
	})

	client := NewClient("u1", NewHub(), nil)

	payload, _ := json.Marshal(SendPayload{ChatID: "c1", Content: "hi"})
	raw, _ := json.Marshal(Envelope{Type: TypeChatSend, Payload: payload})
	router.Route(client, raw)

	rawPing, _ := json.Marshal(Envelope{Type: TypePing})
	router.Route(client, rawPing)

	if len(send.calls) != 1 || send.calls[0].Type != TypeChatSend {
		t.Fatalf("send handler calls = %#v, want 1 TypeChatSend", send.calls)
	}
	if len(ping.calls) != 1 || ping.calls[0].Type != TypePing {
		t.Fatalf("ping handler calls = %#v, want 1 TypePing", ping.calls)
	}
}

func TestRouter_UnknownTypeSendsError(t *testing.T) {
	router := NewWSRouter(map[string]WSHandler{})
	client := NewClient("u1", NewHub(), nil)

	raw, _ := json.Marshal(Envelope{Type: "nope", RequestID: "r1"})
	router.Route(client, raw)

	select {
	case msg := <-client.send:
		var env Envelope
		if err := json.Unmarshal(msg, &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if env.Type != TypeError {
			t.Fatalf("expected error envelope, got %s", env.Type)
		}
		var payload ErrorPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		if payload.Message != "unknown message type" {
			t.Fatalf("message = %q, want unknown message type", payload.Message)
		}
	default:
		t.Fatal("expected error envelope on unknown type")
	}
}

func TestRouter_InvalidJSON(t *testing.T) {
	router := NewWSRouter(map[string]WSHandler{})
	client := NewClient("u1", NewHub(), nil)

	router.Route(client, []byte("not json"))

	select {
	case msg := <-client.send:
		var env Envelope
		if err := json.Unmarshal(msg, &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if env.Type != TypeError {
			t.Fatalf("expected error envelope, got %s", env.Type)
		}
	default:
		t.Fatal("expected error envelope for invalid json")
	}
}
