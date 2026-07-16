package realtimecontract

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// TestClient is a test-friendly client with a buffered send channel.
// The Hub's Send/BroadCast writes to this channel; tests read via Recv.
type TestClient struct {
	UserID     string
	OpenChatID string
	send       chan []byte
	mu         sync.Mutex
	sent       [][]byte
}

func NewTestClient(userID string) *TestClient {
	return &TestClient{UserID: userID, send: make(chan []byte, 256)}
}

// SendChan returns the underlying send channel (used by hub adapters).
func (c *TestClient) SendChan() chan []byte { return c.send }

// Recv blocks until a message is received or timeout expires.
func (c *TestClient) Recv(timeout time.Duration) ([]byte, error) {
	select {
	case msg := <-c.send:
		c.mu.Lock()
		c.sent = append(c.sent, msg)
		c.mu.Unlock()
		return msg, nil
	case <-time.After(timeout):
		return nil, errors.New("timeout")
	}
}

// Sent returns all messages received so far.
func (c *TestClient) Sent() [][]byte { c.mu.Lock(); defer c.mu.Unlock(); return c.sent }

// LastSent returns the most recent message.
func (c *TestClient) LastSent() []byte { c.mu.Lock(); defer c.mu.Unlock(); return last(c.sent) }

// Send puts a message on the send channel (simulates incoming from hub).
func (c *TestClient) Send(msg []byte) { c.send <- msg }

func last(s [][]byte) []byte {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// Message is a realtime chat message payload (decoupled from domain).
type Message struct {
	ID              int
	ChatID          string
	SenderID        string
	Content         string
	CreatedAt       time.Time
	ClientMessageID *string
}

// Hub defines the contract for a realtime connection manager.
// TestClients are created by the test, registered with the Hub,
// and receive messages via their send channel (readable with Recv).
type Hub interface {
	// Register adds a client. Must call Send on all other clients
	// with an "isOnlineStatus.update" broadcast on first connection for a user.
	Register(client *TestClient)
	// Unregister removes a client. Must broadcast offline if last connection.
	Unregister(client *TestClient)
	// Send delivers msg to all connections of toUserID.
	Send(toUserID string, msg []byte)
	// BroadCast delivers msg to all connected clients.
	BroadCast(msg []byte)
	// IsOnline returns true if the user has at least one active connection.
	IsOnline(userID string) bool
	// OnlineUserIDs returns the set of unique connected user IDs.
	OnlineUserIDs() []string
	// OpenChat marks the client as observing a chat room.
	OpenChat(client *TestClient, chatID string)
	// CloseChat removes the client from the chat observer set.
	CloseChat(client *TestClient)
	// GetObserversForChat returns user IDs observing the chat, excluding excludeUserID.
	GetObserversForChat(chatID, excludeUserID string) []string
	// SendToUser sends a formatted chat message envelope to the user's connections.
	SendToUser(toUserID, requestID string, msg *Message)
}

// --- Client lifecycle contract tests ---

func TestRegister_AddsClient(t *testing.T, hub Hub) {
	c := NewTestClient("user-1")
	hub.Register(c)

	if !hub.IsOnline("user-1") {
		t.Error("client should be online after Register")
	}
	ids := hub.OnlineUserIDs()
	if len(ids) != 1 || ids[0] != "user-1" {
		t.Errorf("OnlineUserIDs = %v, want [user-1]", ids)
	}
}

func TestUnregister_RemovesClient(t *testing.T, hub Hub) {
	c := NewTestClient("user-1")
	hub.Register(c)
	hub.Unregister(c)

	if hub.IsOnline("user-1") {
		t.Error("client should be offline after Unregister")
	}
	if len(hub.OnlineUserIDs()) != 0 {
		t.Error("OnlineUserIDs should be empty after Unregister")
	}
}

func TestMultipleClientsSameUser(t *testing.T, hub Hub) {
	c1 := NewTestClient("user-1")
	c2 := NewTestClient("user-1")
	hub.Register(c1)
	hub.Register(c2)

	if !hub.IsOnline("user-1") {
		t.Error("user-1 should be online")
	}

	hub.Unregister(c1)
	if !hub.IsOnline("user-1") {
		t.Error("user-1 should still be online with second connection")
	}

	hub.Unregister(c2)
	if hub.IsOnline("user-1") {
		t.Error("user-1 should be offline after all connections removed")
	}
}

func TestRegister_Unregister_Idempotent(hub Hub) {
	hub.Unregister(NewTestClient("never-registered"))
	c := NewTestClient("user-a")
	hub.Register(c)
	hub.Unregister(NewTestClient("different-instance"))
}

// --- Message delivery contract tests ---

func TestSend_DeliversToSpecificUser(t *testing.T, hub Hub) {
	c := NewTestClient("user-1")
	hub.Register(c)
	drainRecv(c)

	hub.Send("user-1", []byte("hello"))

	msg, err := c.Recv(time.Second)
	if err != nil {
		t.Fatal("should receive message: ", err)
	}
	if string(msg) != "hello" {
		t.Errorf("got %q, want %q", string(msg), "hello")
	}
}

func TestSend_OnlyDeliversToTargetUser(t *testing.T, hub Hub) {
	c1 := NewTestClient("user-1")
	c2 := NewTestClient("user-2")
	hub.Register(c1)
	hub.Register(c2)
	drainRecv(c1)
	drainRecv(c2)

	hub.Send("user-1", []byte("only for user-1"))

	_, err := c1.Recv(time.Second)
	if err != nil {
		t.Error("user-1 should have received a message: ", err)
	}
	_, err = c2.Recv(200 * time.Millisecond)
	if err == nil {
		t.Error("user-2 should NOT have received a message")
	}
}

func TestSend_DeliversToAllConnectionsOfUser(t *testing.T, hub Hub) {
	cA := NewTestClient("user-1")
	cB := NewTestClient("user-1")
	hub.Register(cA)
	hub.Register(cB)
	drainRecv(cA)
	drainRecv(cB)

	hub.Send("user-1", []byte("multi-conn"))

	if _, err := cA.Recv(time.Second); err != nil {
		t.Error("connection A should receive message: ", err)
	}
	if _, err := cB.Recv(time.Second); err != nil {
		t.Error("connection B should receive message: ", err)
	}
}

func TestBroadcast_DeliversToAllUsers(t *testing.T, hub Hub) {
	c1 := NewTestClient("user-1")
	c2 := NewTestClient("user-2")
	hub.Register(c1)
	hub.Register(c2)
	drainRecv(c1)
	drainRecv(c2)

	hub.BroadCast([]byte("broadcast"))

	if _, err := c1.Recv(time.Second); err != nil {
		t.Error("user-1 should receive broadcast: ", err)
	}
	if _, err := c2.Recv(time.Second); err != nil {
		t.Error("user-2 should receive broadcast: ", err)
	}
}

// --- Online/Offline broadcasting contract tests ---

func TestRegister_BroadcastsOnlineStatus(t *testing.T, hub Hub) {
	observer := NewTestClient("observer")
	hub.Register(observer)
	drainRecv(observer)

	newUser := NewTestClient("user-new")
	hub.Register(newUser)

	msg, err := observer.Recv(time.Second)
	if err != nil {
		t.Fatal("observer should receive online broadcast: ", err)
	}
	if !contains(msg, `"isOnline":true`) || !contains(msg, "user-new") {
		t.Errorf("observer should receive online=true for user-new, got: %s", string(msg))
	}
}

func TestUnregister_BroadcastsOfflineStatus(t *testing.T, hub Hub) {
	observer := NewTestClient("observer")
	hub.Register(observer)
	drainRecv(observer)

	gone := NewTestClient("user-gone")
	hub.Register(gone)
	drainRecv(observer)

	hub.Unregister(gone)

	msg, err := observer.Recv(time.Second)
	if err != nil {
		t.Fatal("observer should receive offline broadcast: ", err)
	}
	if !contains(msg, `"isOnline":false`) || !contains(msg, "user-gone") {
		t.Errorf("observer should receive offline=false for user-gone, got: %s", string(msg))
	}
}

// --- Chat observer contract tests ---

func TestOpenChat_TracksObserver(t *testing.T, hub Hub) {
	c := NewTestClient("user-a")
	hub.Register(c)

	observers := hub.GetObserversForChat("chat-1", "other-user")
	if len(observers) != 0 {
		t.Error("no observers should exist before OpenChat")
	}

	hub.OpenChat(c, "chat-1")

	observers = hub.GetObserversForChat("chat-1", "other-user")
	if len(observers) != 1 {
		t.Fatalf("expected 1 observer, got %d", len(observers))
	}
	if observers[0] != "user-a" {
		t.Errorf("expected user-a, got %q", observers[0])
	}

	selfObservers := hub.GetObserversForChat("chat-1", "user-a")
	if len(selfObservers) != 0 {
		t.Error("should not return own user as observer")
	}
}

func TestOpenChat_SwitchesChat(t *testing.T, hub Hub) {
	c := NewTestClient("user-a")
	hub.Register(c)
	hub.OpenChat(c, "chat-1")
	hub.OpenChat(c, "chat-2")

	observers1 := hub.GetObserversForChat("chat-1", "other")
	if len(observers1) != 0 {
		t.Error("should not observe old chat after switching")
	}

	observers2 := hub.GetObserversForChat("chat-2", "other")
	if len(observers2) != 1 {
		t.Fatalf("expected 1 observer on chat-2, got %d", len(observers2))
	}
}

func TestCloseChat_RemovesObserver(t *testing.T, hub Hub) {
	c := NewTestClient("user-a")
	hub.Register(c)
	hub.OpenChat(c, "chat-1")
	hub.CloseChat(c)

	observers := hub.GetObserversForChat("chat-1", "other")
	if len(observers) != 0 {
		t.Error("observer should be removed after CloseChat")
	}
}

func TestCloseChat_Idempotent(hub Hub) {
	c := NewTestClient("user-a")
	hub.Register(c)
	hub.CloseChat(c)
	hub.CloseChat(c)
}

// --- OnlineUserIDs contract tests ---

func TestOnlineUserIDs_ReturnsAllConnectedUsers(t *testing.T, hub Hub) {
	hub.Register(NewTestClient("alice"))
	hub.Register(NewTestClient("bob"))
	hub.Register(NewTestClient("alice"))

	ids := hub.OnlineUserIDs()
	if len(ids) != 2 {
		t.Fatalf("expected 2 unique users, got %d: %v", len(ids), ids)
	}

	m := make(map[string]bool)
	for _, id := range ids {
		m[id] = true
	}
	if !m["alice"] {
		t.Error("expected alice in OnlineUserIDs")
	}
	if !m["bob"] {
		t.Error("expected bob in OnlineUserIDs")
	}
}

func TestIsOnline_MultipleConnections(t *testing.T, hub Hub) {
	hub.Register(NewTestClient("alice"))
	if !hub.IsOnline("alice") {
		t.Error("alice should be online with one connection")
	}

	hub.Register(NewTestClient("alice"))
	if !hub.IsOnline("alice") {
		t.Error("alice should still be online with two connections")
	}
}

// --- SendToUser contract test ---

func TestSendToUser_FormatsChatMessage(t *testing.T, hub Hub) {
	c := NewTestClient("user-b")
	hub.Register(c)
	drainRecv(c)

	now := time.Now()
	msg := &Message{
		ID:              42,
		ChatID:          "chat-1",
		SenderID:        "user-a",
		Content:         "hello",
		CreatedAt:       now,
		ClientMessageID: nil,
	}

	hub.SendToUser("user-b", "req-123", msg)

	recv, err := c.Recv(time.Second)
	if err != nil {
		t.Fatal("should receive message: ", err)
	}
	if !contains(recv, `"type":"chat.message"`) {
		t.Errorf("expected chat.message type in: %s", string(recv))
	}
	if !contains(recv, `"request_id":"req-123"`) {
		t.Errorf("expected request_id in: %s", string(recv))
	}
	if !contains(recv, `"chat_id":"chat-1"`) {
		t.Errorf("expected chat_id in: %s", string(recv))
	}
}

// --- helpers ---

func drainRecv(c *TestClient) {
	for {
		_, err := c.Recv(50 * time.Millisecond)
		if err != nil {
			return
		}
	}
}

func contains(b []byte, s string) bool {
	haystack := string(b)
	for i := 0; i <= len(haystack)-len(s); i++ {
		if haystack[i:i+len(s)] == s {
			return true
		}
	}
	return false
}
