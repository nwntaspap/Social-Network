package ws

import (
	"sync"
	"testing"

	"social-network/internal/core/realtime/realtimecontract"
	"social-network/internal/domain/chat"
)

type hubAdapter struct {
	inner   *Hub
	mu      sync.Mutex
	clients map[*realtimecontract.TestClient]*Client
}

func newHubAdapter() *hubAdapter {
	return &hubAdapter{
		inner:   NewHub(),
		clients: make(map[*realtimecontract.TestClient]*Client),
	}
}

func (a *hubAdapter) getOrCreate(client *realtimecontract.TestClient) *Client {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c, ok := a.clients[client]; ok {
		c.OpenChatId = client.OpenChatID
		return c
	}
	c := &Client{
		UserID:     client.UserID,
		hub:        a.inner,
		conn:       nil,
		send:       client.SendChan(),
		OpenChatId: client.OpenChatID,
	}
	a.clients[client] = c
	return c
}

func (a *hubAdapter) get(client *realtimecontract.TestClient) *Client {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.clients[client]
}

func (a *hubAdapter) Register(client *realtimecontract.TestClient) {
	a.inner.Register(a.getOrCreate(client))
}

func (a *hubAdapter) Unregister(client *realtimecontract.TestClient) {
	realClient := a.get(client)
	if realClient == nil {
		realClient = a.getOrCreate(client)
	}
	a.inner.Unregister(realClient)
}

func (a *hubAdapter) Send(toUserID string, msg []byte) {
	a.inner.Send(toUserID, msg)
}

func (a *hubAdapter) BroadCast(msg []byte) {
	a.inner.BroadCast(msg)
}

func (a *hubAdapter) IsOnline(userID string) bool {
	return a.inner.IsOnline(userID)
}

func (a *hubAdapter) OnlineUserIDs() []string {
	return a.inner.OnlineUserIDs()
}

func (a *hubAdapter) OpenChat(client *realtimecontract.TestClient, chatID string) {
	realClient := a.getOrCreate(client)
	a.inner.OpenChat(realClient, chatID)
	client.OpenChatID = realClient.OpenChatId
}

func (a *hubAdapter) CloseChat(client *realtimecontract.TestClient) {
	realClient := a.getOrCreate(client)
	a.inner.CloseChat(realClient)
	client.OpenChatID = realClient.OpenChatId
}

func (a *hubAdapter) GetObserversForChat(chatID, excludeUserID string) []string {
	clients := a.inner.GetObserversForChat(chatID, excludeUserID)
	ids := make([]string, len(clients))
	for i, c := range clients {
		ids[i] = c.UserID
	}
	return ids
}

func (a *hubAdapter) SendToUser(toUserID, requestID string, msg *realtimecontract.Message) {
	chatMsg := &chat.Message{
		ID:              msg.ID,
		ChatID:          msg.ChatID,
		SenderID:        msg.SenderID,
		Content:         msg.Content,
		CreatedAt:       msg.CreatedAt,
		ClientMessageID: msg.ClientMessageID,
	}
	a.inner.SendToUser(toUserID, requestID, chatMsg)
}

// --- lifecycle ---

func TestInfra_Register_AddsClient(t *testing.T) {
	realtimecontract.TestRegister_AddsClient(t, newHubAdapter())
}

func TestInfra_Unregister_RemovesClient(t *testing.T) {
	realtimecontract.TestUnregister_RemovesClient(t, newHubAdapter())
}

func TestInfra_MultipleClientsSameUser(t *testing.T) {
	realtimecontract.TestMultipleClientsSameUser(t, newHubAdapter())
}

func TestInfra_Register_Unregister_Idempotent(t *testing.T) {
	realtimecontract.TestRegister_Unregister_Idempotent(newHubAdapter())
}

// --- message delivery ---

func TestInfra_Send_DeliversToSpecificUser(t *testing.T) {
	realtimecontract.TestSend_DeliversToSpecificUser(t, newHubAdapter())
}

func TestInfra_Send_OnlyDeliversToTargetUser(t *testing.T) {
	realtimecontract.TestSend_OnlyDeliversToTargetUser(t, newHubAdapter())
}

func TestInfra_Send_DeliversToAllConnectionsOfUser(t *testing.T) {
	realtimecontract.TestSend_DeliversToAllConnectionsOfUser(t, newHubAdapter())
}

func TestInfra_Broadcast_DeliversToAllUsers(t *testing.T) {
	realtimecontract.TestBroadcast_DeliversToAllUsers(t, newHubAdapter())
}

// --- status broadcasting ---

func TestInfra_Register_BroadcastsOnlineStatus(t *testing.T) {
	realtimecontract.TestRegister_BroadcastsOnlineStatus(t, newHubAdapter())
}

func TestInfra_Unregister_BroadcastsOfflineStatus(t *testing.T) {
	realtimecontract.TestUnregister_BroadcastsOfflineStatus(t, newHubAdapter())
}

// --- chat observers ---

func TestInfra_OpenChat_TracksObserver(t *testing.T) {
	realtimecontract.TestOpenChat_TracksObserver(t, newHubAdapter())
}

func TestInfra_OpenChat_SwitchesChat(t *testing.T) {
	realtimecontract.TestOpenChat_SwitchesChat(t, newHubAdapter())
}

func TestInfra_CloseChat_RemovesObserver(t *testing.T) {
	realtimecontract.TestCloseChat_RemovesObserver(t, newHubAdapter())
}

func TestInfra_CloseChat_Idempotent(t *testing.T) {
	realtimecontract.TestCloseChat_Idempotent(newHubAdapter())
}

// --- OnlineUserIDs ---

func TestInfra_OnlineUserIDs_ReturnsAllConnectedUsers(t *testing.T) {
	realtimecontract.TestOnlineUserIDs_ReturnsAllConnectedUsers(t, newHubAdapter())
}

func TestInfra_IsOnline_MultipleConnections(t *testing.T) {
	realtimecontract.TestIsOnline_MultipleConnections(t, newHubAdapter())
}

// --- SendToUser ---

func TestInfra_SendToUser_FormatsChatMessage(t *testing.T) {
	realtimecontract.TestSendToUser_FormatsChatMessage(t, newHubAdapter())
}
