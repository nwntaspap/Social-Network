package realtime

import (
	"encoding/json"
	"maps"
)

type WSRouter interface {
	Route(client *Client, raw []byte)
}

type WSHandler interface {
	Handle(client *Client, env Envelope)
}

// HandlerFunc adapts a function to the WSHandler interface.
type HandlerFunc func(client *Client, env Envelope)

func (f HandlerFunc) Handle(client *Client, env Envelope) {
	f(client, env)
}

type wsRouter struct {
	handlers map[string]WSHandler
}

// NewWSRouter builds a router for the given message type->handler pairs.
// Types without a registered handler respond with an error envelope.
func NewWSRouter(pairs map[string]WSHandler) WSRouter {
	r := &wsRouter{handlers: make(map[string]WSHandler, len(pairs))}
	maps.Copy(r.handlers, pairs)
	return r
}

func (r *wsRouter) Route(client *Client, raw []byte) {
	var env Envelope
	err := json.Unmarshal(raw, &env)
	if err != nil {
		sendError(client, "", "invalid message format")
		return
	}

	handler, ok := r.handlers[env.Type]
	if !ok {
		sendError(client, env.RequestID, "unknown message type")
		return
	}
	handler.Handle(client, env)
}

func sendError(client *Client, requestID, message string) {
	payload, _ := json.Marshal(ErrorPayload{Message: message})
	reply, _ := json.Marshal(Envelope{
		Type:      TypeError,
		RequestID: requestID,
		Payload:   payload,
	})
	client.send <- reply
}
