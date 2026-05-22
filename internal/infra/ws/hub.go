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
	mu            sync.RWMutex
	clients       map[string]map[*Client]bool //userID --> set of connections.
	chatObservers map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[string]map[*Client]bool),
		chatObservers: make(map[string]map[*Client]bool),
	}
}

// register adds a client connection for a user.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()

	becameOnline := h.clients[client.UserID] == nil
	if becameOnline {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true

	log.Printf("ws: user %s connected (%d connections)", client.UserID, len(h.clients[client.UserID]))
	h.mu.Unlock()

	if becameOnline {
		h.BroadCastIsOnlineStatus(client.UserID, true)
	}
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()

	conns, ok := h.clients[client.UserID]
	if !ok {
		h.mu.Unlock()
		return
	}

	if client.OpenChatId != "" {
		delete(h.chatObservers[client.OpenChatId], client)
	}
	if len(h.chatObservers[client.OpenChatId]) == 0 {
		delete(h.chatObservers, client.OpenChatId)
	}
	delete(conns, client)
	becameOffline := len(conns) == 0
	if becameOffline {
		delete(h.clients, client.UserID)
	}

	log.Printf("ws: user %s disconnected", client.UserID)
	h.mu.Unlock()

	if becameOffline {
		h.BroadCastIsOnlineStatus(client.UserID, false)
	}
}

func (h *Hub) OpenChat(client *Client, chatID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.OpenChatId == chatID {
		return
	}

	if client.OpenChatId != "" {
		oldChatID := client.OpenChatId
		delete(h.chatObservers[oldChatID], client)
		if len(h.chatObservers[oldChatID]) == 0 {
			delete(h.chatObservers, oldChatID)
		}
	}

	if h.chatObservers[chatID] == nil {
		h.chatObservers[chatID] = make(map[*Client]bool)
	}

	h.chatObservers[chatID][client] = true
	client.OpenChatId = chatID
}

func (h *Hub) CloseChat(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.OpenChatId == "" {
		return
	}

	oldChatID := client.OpenChatId
	client.OpenChatId = ""
	delete(h.chatObservers[oldChatID], client)
	if len(h.chatObservers[oldChatID]) == 0 {
		delete(h.chatObservers, oldChatID)
	}
}

// returns all the client connections which have currently the chat with chatID open
// besides the the client that is currently typing with ownUserID
func (h *Hub) GetObserversForChat(chatID, ownUserID string) (observers []*Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for k := range h.chatObservers[chatID] {
		if k.UserID != ownUserID {
			observers = append(observers, k)
		}
	}

	return observers
}

// Send delivers a message to all connections of a specific user.
// client.send is a channel, with a specific buffer, which can block if the channel is full
// sending to a channel will wait for as long the channel is full
// keeping the hubs mutex locked for this period is dangerous
// some hub operations(register/unregister) might freeze because of the lock
// for this a copry of the recipients is made and then sends to every recepient
func (h *Hub) Send(toUserID string, msg []byte) {
	h.mu.RLock()
	//these are the recepients
	clients := make([]*Client, 0)
	for client := range h.clients[toUserID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	//sending to every recepient
	for _, client := range clients {
		client.send <- msg
	}
}

func (h *Hub) BroadCast(msg []byte) {
	h.mu.RLock()
	clients := make([]*Client, 0)
	for _, conns := range h.clients {
		for client := range conns {
			clients = append(clients, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range clients {
		client.send <- msg
	}
}
func (h *Hub) BroadCastIsOnlineStatus(userID string, isOnline bool) {
	outPayload, _ := json.Marshal(IsOnlineStatusPayload{
		UserID:   userID,
		IsOnline: isOnline,
	})
	statusToBeBroadcasted, _ := json.Marshal(Envelope{
		Type:    TypeIsOnlineStatus,
		Payload: outPayload,
	})
	h.BroadCast(statusToBeBroadcasted)
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
	clients := make([]*Client, 0)
	for client := range h.clients[toUserID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()
	for _, client := range clients {
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
