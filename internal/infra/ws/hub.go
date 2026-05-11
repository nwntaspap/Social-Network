package ws

import (
	"encoding/json"
	"github.com/arnald/forum/internal/domain/chat"
	"log"
	"sync"
)

// Hub manages all active WebSocket Connections.
// One Hub runs for the lifetime of the server.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]bool //userID --> set of connections.
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]bool),
	}
}

// register adds a client connection for a user.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true

	log.Printf("ws: user %s connected (%d connections)", client.UserID, len(h.clients[client.UserID]))
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns, ok := h.clients[client.UserID]
	if !ok {
		return
	}

	delete(conns, client)
	if len(conns) == 0 {
		delete(h.clients, client.UserID)
	}

	log.Printf("ws: user %s disconnected", client.UserID)
}

// Send delivers a message to all connections of a specific user.
func (h *Hub) Send(toUserID string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients[toUserID] {
		client.send <- msg
	}
}

// this is here so that the hub ipmlement the broadcaster interface in the app
// for this, the chat domain is imported, if we dedide we dont want that the next 2 methods should be implemented for a different struct so that the hub remains ws generic and not chat specific
func (h *Hub) SendToUser(toUserID, requestID string, msg *chat.Message) {
	outPayload, _ := json.Marshal(MessagePayload{
		ID:              msg.ID,
		ChatID:          msg.ChatID,
		SenderID:        msg.SenderID,
		Content:         msg.Content,
		CreatedAt:       msg.CreatedAt,
		ClientMessageID: msg.ClientMessageID,
	})
	reply, _ := json.Marshal(Envelope{
		Type:      TypeChatMessage,
		RequestID: requestID,
		Payload:   outPayload,
	})
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients[toUserID] {
		client.send <- reply
	}
}

// IsOnline returns true if a user has at least one active connection.
func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients[userID]) > 0
}

// OnlineUserIDs returns the set of currently connected user IDs.
func (h *Hub) OnlineUserIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}
